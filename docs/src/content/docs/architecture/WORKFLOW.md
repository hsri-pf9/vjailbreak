# vJailbreak Detailed Workflow

This document explains how vJailbreak (VJB) orchestrates end‑to‑end VMware → OpenStack migrations, with emphasis on:

- How flavors are selected for target VMs in OpenStack
- How source VMs are attached to the VJB helper VM for conversion, then detached
- How the final OpenStack VM is created from the converted image/volumes

It complements the high‑level features covered in `README.md` and focuses on actionable, operator‑relevant details for deployers.

## Components and Roles

- **VJB Controller (Kubernetes operators/controllers)**
  - Reconciles CRDs like `MigrationPlan`, `MigrationTemplate`, `NetworkMapping`, `StorageMapping`, `VMwareCreds`, and `OpenStackCreds` under `k8s/migration/`.
  - Plans migrations, manages job orchestration, tracks status, runs pre/post checks.

- **Resource Manager (`resmgr`)**
  - Discovers VMware inventory, flavors, networks, storage targets; resolves mappings.
  - Path: `k8s/migration/pkg/sdk/resmgr/resmgr.go`.

- **v2v-helper**
  - A containerized helper that runs `virt-v2v` (2.7.13) and related tooling to convert VMware disks to OpenStack‑compatible images/volumes.
  - See `v2v-helper/` and `v2v-helper/vm/vmops.go` for VM/disk operations.

- **VMware/vSphere**
  - Source ESXi hosts and VMs discovered via vCenter APIs.

- **OpenStack**
  - Destination clouds: Glance (images), Cinder (volumes), Neutron (networks), Nova (compute flavors/instances).

## Migration Phase and State Model

The controllers surface per‑VM progress using constants in `constants.go`:

- __Phases__: `VMMigrationPhasePending`, `Validating`, `AwaitingDataCopyStart`, `Copying`, `CopyingChangedBlocks`, `ConvertingDisk`, `AwaitingCutOverStartTime`, `AwaitingAdminCutOver`, `Succeeded`, `Failed`, `Unknown` (see `VMMigrationStatesEnum`).
- __Pod Conditions__: `MigrationConditionTypeDataCopy`, `Migrating`, `Validated`, `Failed` reflect job‑level status.

Operators will see these in CR status and controller logs to trace where a migration is in the pipeline.

### Phase-by-Phase Flow

- **Phase 1: Pending**
  - Trigger: `MigrationPlan` created/updated; VM enqueued.
  - Action: Basic validation and inventory discovery.
  - Next: `Validating`.

- **Phase 2: Validating**
  - Action: Check credentials reachability, `NetworkMapping` and `StorageMapping` existence/capacity, quotas, and policy.
  - Pod condition: `Validated` on success, `Failed` on error.
  - Next: `AwaitingDataCopyStart` if conversion is authorized; else `Failed`.

- **Phase 3: AwaitingDataCopyStart**
  - Action: Schedule/helper prep. Ensure snapshot policy, CBT for hot path using `IsCBTEnabled()`/`EnableCBT()`; then `TakeSnapshot()`.
  - Next: `Copying` when the conversion job starts.

- **Phase 4: Copying**
  - Action: Initial bulk copy. `UpdateDisksInfo()` records snapshot backing and ChangeIDs. `virt-v2v` converts system/data disks to artifacts.
  - Pod condition: `DataCopy`.
  - Next: Hot path → `CopyingChangedBlocks`; Cold path → `ConvertingDisk` or straight to spawn depending on pipeline wiring.

- **Phase 5: CopyingChangedBlocks** (hot only)
  - Action: Delta copy using VMware ChangeIDs via `CustomQueryChangedDiskAreas()`; optional quiesce window.
  - Next: `ConvertingDisk` after final sync, or directly to spawn if conversion integrated.

- **Phase 6: ConvertingDisk**
  - Action: Finalization of converted artifacts, driver/boot remediation, upload to Cinder/Glance.
  - Next: `AwaitingCutOverStartTime` (scheduled) or `AwaitingAdminCutOver` (manual gate) based on plan policy; else directly to spawn in fully automated mode.

- **Phase 7: AwaitingCutOverStartTime**
  - Action: Wait until planned cutover timestamp; controllers use backoff params `VMActiveWaitIntervalSeconds`/`VMActiveWaitRetryLimit` for polling cadence elsewhere.
  - Next: `AwaitingAdminCutOver` or proceed to spawn.

- **Phase 8: AwaitingAdminCutOver**
  - Action: Human approval window. On approval, proceed with source shutdown (`VMGuestShutdown()` or `VMPowerOff()`) and networking safeguards (`DisconnectNetworkInterfaces()` if enabled).
  - Next: Target VM spawn.

- **Phase 9: Succeeded**
  - Action: Target VM ACTIVE, health checks pass; cleanup snapshots via `DeleteSnapshot*`/`DeleteMigrationSnapshots()`/`CleanUpSnapshots()`; optional ESXi host removal.

- **Phase 10: Failed**
  - Action: Error recorded with reason; retry/backoff depending on failure class. Idempotent steps allow re‑runs.

Pod Conditions mapping (typical):
- `DataCopy` → bulk copy in progress.
- `Migrating` → conversion/spawn in progress.
- `Validated` → prechecks passed.
- `Failed` → terminal or retriable error.

```mermaid
stateDiagram-v2
    [*] --> Pending
    Pending --> Validating
    Validating --> AwaitingDataCopyStart: ok
    Validating --> Failed: error
    AwaitingDataCopyStart --> Copying
    Copying --> CopyingChangedBlocks: hot
    Copying --> ConvertingDisk: cold
    CopyingChangedBlocks --> ConvertingDisk
    ConvertingDisk --> AwaitingCutOverStartTime: scheduled
    ConvertingDisk --> AwaitingAdminCutOver: manual
    ConvertingDisk --> Spawn: auto
    AwaitingCutOverStartTime --> AwaitingAdminCutOver
    AwaitingAdminCutOver --> Spawn
    Spawn --> Succeeded
    Spawn --> Failed
    Failed --> [*]
    Succeeded --> [*]
```

## Detailed Mechanics: Attach → Data Copy → Convert → Detach → Spawn

1. __Attach (Source Presentation)__
   - Introspect source VM to capture CPU, memory, disks, firmware (UEFI detection).
   - Ensure CBT (Changed Block Tracking) if hot migration is configured:
     - Check if CBT is enabled; if not, enable CBT followed by a fresh snapshot.
   - Establish a consistent view of disks:
     - Create snapshot for cold/hot initial pass.
     - Populate per‑disk snapshot metadata and VMware ChangeIDs.
   - Networking safeguards (optional policy): disconnect network interfaces prior to cutover to avoid IP conflicts.

2. __Data Copy (Initial Transfer)__
   - Initial bulk copy of VM disk data from VMware to staging area or directly to OpenStack storage.
   - Record snapshot backing and VMware ChangeIDs for tracking.
   - For hot migrations: continuous monitoring of changed blocks.
   - Data integrity verification during transfer process.

3. __Convert (virt‑v2v Execution)__
   - Disks exposed to the helper are read by virt-v2v to produce raw/qcow2 format.
   - System remediation performed by virt‑v2v (virtio drivers, bootloader, fstab, cloud‑init enablement) based on detected OS family.
   - For hot delta runs, final changed regions are derived using VMware ChangeIDs recorded per disk.

4. __Detach (Source Cleanup)__
   - On successful conversion, helper disconnects from export path and removes temporary constructs.
   - Snapshot lifecycle:
     - Point deletions and bulk cleanup of migration snapshots.
   - For cold migrations, final consolidation is performed by vCenter on snapshot removal.

5. __Spawn (Target VM Creation)__
   - Storage:
     - Converted artifacts go to Cinder for boot/data volumes; Glance images used when policy prefers image‑based boot.
     - RDM disks (when present) carry backend hints to map to specific Cinder volume types/pools.
   - Networking:
     - Create Neutron ports per NIC mapped from VMware PG/VLAN using NetworkMapping; attach to the new server.
   - Compute:
     - Pick flavor (see Flavor Selection Algorithm section) and create Nova server with appropriate backoff parameters.

```mermaid
sequenceDiagram
    participant VC as vCenter/ESXi
    participant VJB as VJB Controller
    participant H as V2V Helper
    participant OS as OpenStack (Glance/Cinder/Neutron/Nova)

    VJB->>VC: Inspect VM
    VJB->>VC: Ensure CBT + Snapshot
    VJB->>H: Start conversion job with export endpoints
    H->>VC: Read disks and track changes
    H->>H: Run virt-v2v convert
    H->>OS: Upload to Cinder/Glance
    H-->>VC: Detach, cleanup snapshots
    VJB->>OS: Create Neutron ports + Nova server (chosen flavor)
    OS-->>VJB: VM ACTIVE, IP assigned
```

## Target Provisioning (OpenStack)


### Storage (Cinder/Glance)

- The helper creates volumes sized from converted disks (+ optional growth buffer).
- Boot strategy:
  - Boot‑from‑volume (recommended): root volume created from the converted artifact; bypasses flavor root disk limits.
  - Image‑based: upload to Glance, then Nova creates a boot disk from the image.
- Volume types/backends are selected from StorageMapping; RDM hints (pool, type) propagate to volume create options.

### Networking (Neutron Ports)

- Each VMware NIC maps to a Neutron network from NetworkMapping.
- For each NIC, create a port with security groups/port‑security policies, then attach to the Nova server in create.
- MAC preservation is optional and policy‑driven; default is to re‑DHCP via cloud‑init.

## Observability, Reliability, Security

- __Phases and conditions__: update using `VMMigrationStatesEnum`, `MigrationConditionType*` at each step.
- __Retries/backoff__: transient API failures retried; long waits use `VMActiveWaitIntervalSeconds`, `VMActiveWaitRetryLimit`.
- __Idempotency__: detect/reuse existing images/volumes for safe re‑runs.
- __Secrets__: store credentials in Kubernetes Secrets; least-privilege access.
- __Encrypted data plane__: HTTPS for Glance/Cinder; TLS for NBD where enabled.
- __Snapshot/export hygiene__: snapshots/exports are ephemeral, cleaned up promptly, and staging locations are access‑controlled.

## Cleanup and Decommission

- __Snapshot cleanup__: remove migration snapshots via `DeleteSnapshot*`, `DeleteMigrationSnapshots()` or `CleanUpSnapshots()` after success/finalization.
- __ESXi host decommission (optional)__: when a host is drained and in maintenance mode, the controller can remove it from vCenter per policy for clean source decommissioning.


## End-to-End Flow

1. Discovery and Planning
   - VJB reads VMware inventory (VM hardware, CPU/RAM, disks, NICs, VLAN/PG, power state, VMware Tools) and OpenStack catalogs (flavors, networks, images, storage backends).
   - Operator prepares `NetworkMapping` and `StorageMapping` CRs to map VMware networks/storage to OpenStack equivalents.
   - `MigrationPlan` references credentials and mappings, and defines the migration mode (cold/hot), batches, and scheduling.

2. Pre-Checks
   - Validates credentials reachability and permissions.
   - Confirms mapped networks and storage exist and have capacity.
   - Optionally places ESXi host/VM into a compliant state (snapshots/CBT as configured).

3. Helper VM Preparation (Attach Phase)
   - VJB brings up a temporary VJB Helper VM (or pod with nested VM tooling) that runs `v2v-helper`.
   - Source VM disks are presented to the helper via one of the supported paths:
     - NFC/VDDK data path, or
     - NBD/iSCSI export from ESXi, or
     - VMDK export to a staging datastore, then mounted by the helper.
   - The helper “attaches” to the source by opening the exported disks (read‑only for cold, or with change capture for hot if configured).

4. Conversion (virt-v2v)
   - `virt-v2v` converts the attached VMDKs to QCOW2/raw suitable for OpenStack.
   - Drivers/services are injected as needed (virtio, cloud‑init, udev rules, fstab fixes, bootloader adjustments).
   - Output can be:
     - Imported into Glance as images, or
     - Written directly to Cinder volumes (preferred for large disks).

5. Target Provisioning (Detach + Spawn New VM)
   - After conversion, the helper detaches from the source exports and tears down transient connections.
   - OpenStack resources are created:
     - Glance image or Cinder volumes per disk according to `StorageMapping` policy.
     - Neutron ports created per NIC according to `NetworkMapping`.
     - Nova server is created with the selected flavor (see Flavor Selection below), attached to the ports and boot volumes.
   - Cloud‑init/user‑data can be injected for first‑boot customization.

6. Cutover and Validation
   - For cold migration: the VM was powered off during conversion; the OpenStack VM is booted and validated.
   - For hot migration: a delta pass may run before cutover; after final sync, source is quiesced and target is booted.
   - Health checks: IP acquisition, SSH/WinRM, application/service probes as configured.

7. Cleanup
   - Temporary helper artifacts removed.
   - Optionally, when an ESXi host has no remaining VMs and is in maintenance mode, VJB invokes safe vCenter removal of the host (configurable policy).

## Flavor Selection Algorithm

VJB aims to preserve (or minimally upscale) source VM sizing while adhering to destination constraints.

Inputs collected from VMware:
- vCPU count, CPU reservation/shares
- Memory size (MiB)
- Disk layout (system vs data, sizes, thin/thick)
- Special features (NUMA, EFI/BIOS, CPU flags) when detectable

OpenStack catalogs:
- List of flavors: vCPUs, RAM, ephemeral/root disk, extra_specs (NUMA, CPU policy)
- Project quotas and host aggregates/availability zones

Selection steps:
1. Baseline Match
   - Find flavors with `vcpus >= source.vcpus` and `ram_mb >= ceil(source.ram_mb)`. If `root_disk_gb` is enforced, ensure it fits the boot volume/image (or prefer boot‑from‑volume to decouple from flavor disk).

2. Policy Adjustments
   - Apply placement policies from `MigrationPlan` (e.g., min upscale, max cap, NUMA affinity, CPU pinning if required by workload).
   - If source has high CPU reservation or latency‑sensitive flags, prefer flavors with `extra_specs` for dedicated CPU or pinned policy.

3. Storage Strategy Interaction
   - If boot‑from‑volume is used (recommended), ignore flavor root disk and size the Cinder volume(s) from converted disk sizes plus growth buffer (configured in `StorageMapping`).

4. Network/Quota Constraints
   - Ensure selected AZ/aggregate has capacity and mapped networks are reachable.

5. Fallbacks
   - If no exact match, pick the smallest flavor that satisfies both CPU and RAM (upscale). If policies disallow upscale, surface a validation error in status and mark plan for manual override.

Overrides:
- Operators can pin a specific flavor per VM or per template via `MigrationTemplate`.
- Per‑VM exceptions (e.g., memory‑heavy DBs) can be annotated in the plan.

Outcome:
- The chosen flavor name/ID is recorded in `MigrationPlan` status for traceability and auditing.


## Hot vs Cold Migration

- Cold (recommended for simplicity)
  - Power off source during conversion.
  - Single pass; minimal risk of data inconsistency.

- Hot (reduced downtime)
  - Initial sync while source is running, then delta pass with quiesce at cutover.
  - Requires CBT/delta export support and appropriate VMware privileges.

## Failure Handling and Observability

- Each phase updates CR status with progress, per‑VM logs, and error details.
- Retries with exponential backoff for transient API/transport failures.
- Idempotent steps for safe re‑runs (e.g., existing images/volumes detection).

## Post-Migration Cleanup

- Temporary helper artifacts removed.
- Optionally, when an ESXi host has no remaining VMs and is in maintenance mode, VJB invokes safe vCenter removal of the host (configurable policy).

## Operator Tips
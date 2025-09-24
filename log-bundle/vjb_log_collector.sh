#!/bin/bash
set -e

LOG_DIR="/var/log/pf9"
OUTPUT_BASE="/tmp/vjb_logs"
TS=$(date +"%Y%m%d_%H%M%S")
OUTPUT_DIR="${OUTPUT_BASE}/${TS}"

mkdir -p "$OUTPUT_DIR"

usage() {
    echo "Usage: $0 --vms <vm1,vm2,...>"
    exit 1
}

normalize_vm_name() {
    echo "$1" | tr -d '.'
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --vms)
            IFS=',' read -r -a VMS <<< "$2"
            shift 2
            ;;
        *)
            usage
            ;;
    esac
done

if [ -z "${VMS[*]}" ]; then
    usage
fi

for VM_RAW in "${VMS[@]}"; do
    VM=$(normalize_vm_name "$VM_RAW")
    echo "[INFO] Processing VM: $VM_RAW (normalized: $VM)"

    FILE_FOUND=false
    VM_OUT="${OUTPUT_DIR}/${VM_RAW}"

    if compgen -G "${LOG_DIR}/migration-${VM}*.log" > /dev/null; then
        mkdir -p "$VM_OUT"
        flat_matches=(${LOG_DIR}/migration-${VM}*.log)
        cp "${flat_matches[@]}" "$VM_OUT/" 2>/dev/null || true
        FILE_FOUND=true
        echo "[INFO]   -> Copied flat log(s)"
    fi

    if compgen -G "${LOG_DIR}/migration-${VM}*/" > /dev/null; then
        mkdir -p "$VM_OUT"
        dir_matches=(${LOG_DIR}/migration-${VM}*/)
        cp -r "${dir_matches[@]}" "$VM_OUT/" 2>/dev/null || true
        FILE_FOUND=true
        echo "[INFO]   -> Copied folder log(s)"
    fi

    if [ "$FILE_FOUND" = false ]; then
        echo "[WARN] No logs found for VM: $VM_RAW"
        rm -rf "$VM_OUT"
    fi
done

if [ "$(ls -A "$OUTPUT_DIR")" ]; then
    tar -czf "${OUTPUT_BASE}/vjb_logs_${TS}.tar.gz" -C "$OUTPUT_BASE" "$TS"
    echo "[INFO] Logs packaged at: ${OUTPUT_BASE}/vjb_logs_${TS}.tar.gz"
else
    echo "[WARN] No logs collected, nothing to package."
    rmdir "$OUTPUT_DIR"
fi


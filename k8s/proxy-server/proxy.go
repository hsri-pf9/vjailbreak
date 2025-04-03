package main

import (
	"context"
	"sigs.k8s.io/apiserver-runtime/pkg/builder"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/rest"
)

// VM represents a virtual machine
type VM struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VMSpec   `json:"spec,omitempty"`
	Status VMStatus `json:"status,omitempty"`
}

// New creates a new VM
func (v *VM) New() runtime.Object {
	return &VM{}
}

// NewList creates a new VM list
func (v *VM) NewList() runtime.Object {
	return &VMList{}
}

// GetGroupVersionResource returns the GroupVersionResource for this resource
func (v *VM) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "vjailbreak.platform9.com",
		Version:  "v1alpha1",
		Resource: "vms",
	}
}

// IsStorageVersion returns true as this is the storage version
func (v *VM) IsStorageVersion() bool {
	return true
}

// GetObjectMeta returns the object meta reference
func (v *VM) GetObjectMeta() *metav1.ObjectMeta {
	return &v.ObjectMeta
}

// NamespaceScoped returns true as VMs are namespaced
func (v *VM) NamespaceScoped() bool {
	return true
}

// ShortNames returns the short names for this resource
func (v *VM) ShortNames() []string {
	return []string{"vm"}
}

// GetStatus returns the VM status
func (v *VM) GetStatus() interface{} {
	return v.Status
}

// SetStatus sets the VM status
func (v *VM) SetStatus(status interface{}) {
	v.Status = status.(VMStatus)
}

// GetSpec returns the VM spec
func (v *VM) GetSpec() interface{} {
	return v.Spec
}

// SetSpec sets the VM spec
func (v *VM) SetSpec(spec interface{}) {
	v.Spec = spec.(VMSpec)
}

// GetValidationSchema returns the validation schema for this resource
func (v *VM) GetValidationSchema() rest.ValidationSchema {
	return rest.ValidationSchema{
		OpenAPIV3Schema: &runtime.RawExtension{
			Object: nil,
		},
	}
}

// DeepCopyObject returns a deep copy
func (v *VM) DeepCopyObject() runtime.Object {
	if v == nil {
		return nil
	}
	copied := &VM{}
	v.DeepCopyInto(copied)
	return copied
}

// DeepCopyInto copies all properties into another VM
func (v *VM) DeepCopyInto(out *VM) {
	*out = *v
	out.TypeMeta = v.TypeMeta
	v.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = v.Spec
	out.Status = v.Status
}

// GetObjectMeta returns the object meta reference.
func (v *VM) GetObjectMeta() *metav1.ObjectMeta {
	return &v.ObjectMeta
}

// DeepCopyObject returns a deep copy
func (v *VM) DeepCopyObject() runtime.Object {
	if v == nil {
		return nil
	}
	copied := &VM{}
	v.DeepCopyInto(copied)
	return copied
}

// DeepCopyInto copies all properties into another VM
func (v *VM) DeepCopyInto(out *VM) {
	*out = *v
	out.TypeMeta = v.TypeMeta
	v.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = v.Spec
	out.Status = v.Status
}

// NewVM creates a new VM
func NewVM() resource.Object {
	return &VM{}
}

// NewVMList creates a new VM list
func NewVMList() resource.ObjectList {
	return &VMList{}
}

// GetGroupVersionResource returns the GVR for this API
func (v *VM) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "vjailbreak.platform9.com",
		Version:  "v1alpha1",
		Resource: "vms",
	}
}

// IsStorageVersion returns true since this is the storage version
func (v *VM) IsStorageVersion() bool {
	return true
}

// NamespaceScoped returns true since VMs are namespaced
func (v *VM) NamespaceScoped() bool {
	return true
}

type VMSpec struct {
	CPUs     int `json:"cpus"`
	MemoryMB int `json:"memoryMB"`
}

type VMStatus struct {
	State     string `json:"state"`
	IPAddress string `json:"ipAddress,omitempty"`
}

// VMList contains a list of VMs
type VMList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VM `json:"items"`
}

// DeepCopyObject returns a deep copy
func (v *VMList) DeepCopyObject() runtime.Object {
	if v == nil {
		return nil
	}
	copied := &VMList{}
	v.DeepCopyInto(copied)
	return copied
}

// DeepCopyInto copies all properties into another VMList
func (v *VMList) DeepCopyInto(out *VMList) {
	*out = *v
	out.TypeMeta = v.TypeMeta
	v.ListMeta.DeepCopyInto(&out.ListMeta)
	if v.Items != nil {
		out.Items = make([]VM, len(v.Items))
		for i := range v.Items {
			v.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

// VMStorage implements storage for VMs
type VMStorage struct {
	mutex sync.RWMutex
	vms   map[string]*VM
}

func NewVMStorage() *VMStorage {
	return &VMStorage{
		vms: make(map[string]*VM),
	}
}

func (s *VMStorage) Get(name string) (*VM, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	vm, exists := s.vms[name]
	if !exists {
		return nil, fmt.Errorf("VM %s not found", name)
	}
	return vm, nil
}

func (s *VMStorage) List() []VM {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	vms := make([]VM, 0, len(s.vms))
	for _, vm := range s.vms {
		vms = append(vms, *vm)
	}
	return vms
}

func (s *VMStorage) Create(vm *VM) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.vms[vm.Name]; exists {
		return fmt.Errorf("VM %s already exists", vm.Name)
	}

	s.vms[vm.Name] = vm
	return nil
}

func (s *VMStorage) Update(vm *VM) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.vms[vm.Name]; !exists {
		return fmt.Errorf("VM %s not found", vm.Name)
	}

	s.vms[vm.Name] = vm
	return nil
}

func (s *VMStorage) Delete(name string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.vms[name]; !exists {
		return fmt.Errorf("VM %s not found", name)
	}

	delete(s.vms, name)
	return nil
}

package main

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestVM_GetObjectMeta(t *testing.T) {
	vm := &VM{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vm",
			Namespace: "default",
		},
	}

	meta := vm.GetObjectMeta()
	if meta.Name != "test-vm" {
		t.Errorf("Expected name 'test-vm', got '%s'", meta.Name)
	}
	if meta.Namespace != "default" {
		t.Errorf("Expected namespace 'default', got '%s'", meta.Namespace)
	}
}

func TestVM_GetGroupVersionResource(t *testing.T) {
	vm := &VM{}
	gvr := vm.GetGroupVersionResource()
	expected := schema.GroupVersionResource{
		Group:    "vjailbreak.platform9.com",
		Version:  "v1alpha1",
		Resource: "vms",
	}

	if gvr != expected {
		t.Errorf("Expected GVR %v, got %v", expected, gvr)
	}
}

func TestVM_DeepCopy(t *testing.T) {
	original := &VM{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vm",
			Namespace: "default",
		},
		Spec: VMSpec{
			CPUs:     2,
			MemoryMB: 4096,
		},
		Status: VMStatus{
			State:     "running",
			IPAddress: "192.168.1.100",
		},
	}

	copy := original.DeepCopyObject().(*VM)

	// Test that it's a deep copy
	if copy == original {
		t.Error("DeepCopy returned same pointer")
	}

	// Test all fields are copied correctly
	if copy.Name != original.Name {
		t.Errorf("Expected name '%s', got '%s'", original.Name, copy.Name)
	}
	if copy.Namespace != original.Namespace {
		t.Errorf("Expected namespace '%s', got '%s'", original.Namespace, copy.Namespace)
	}
	if copy.Spec.CPUs != original.Spec.CPUs {
		t.Errorf("Expected CPUs %d, got %d", original.Spec.CPUs, copy.Spec.CPUs)
	}
	if copy.Spec.MemoryMB != original.Spec.MemoryMB {
		t.Errorf("Expected MemoryMB %d, got %d", original.Spec.MemoryMB, copy.Spec.MemoryMB)
	}
	if copy.Status.State != original.Status.State {
		t.Errorf("Expected State '%s', got '%s'", original.Status.State, copy.Status.State)
	}
	if copy.Status.IPAddress != original.Status.IPAddress {
		t.Errorf("Expected IPAddress '%s', got '%s'", original.Status.IPAddress, copy.Status.IPAddress)
	}
}

func TestVMList_DeepCopy(t *testing.T) {
	original := &VMList{
		Items: []VM{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm1",
					Namespace: "default",
				},
				Spec: VMSpec{
					CPUs:     2,
					MemoryMB: 4096,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm2",
					Namespace: "default",
				},
				Spec: VMSpec{
					CPUs:     4,
					MemoryMB: 8192,
				},
			},
		},
	}

	copy := original.DeepCopyObject().(*VMList)

	// Test that it's a deep copy
	if copy == original {
		t.Error("DeepCopy returned same pointer")
	}

	// Test that the items are copied correctly
	if len(copy.Items) != len(original.Items) {
		t.Errorf("Expected %d items, got %d", len(original.Items), len(copy.Items))
	}

	for i := range original.Items {
		if copy.Items[i].Name != original.Items[i].Name {
			t.Errorf("Item %d: Expected name '%s', got '%s'", i, original.Items[i].Name, copy.Items[i].Name)
		}
		if copy.Items[i].Spec.CPUs != original.Items[i].Spec.CPUs {
			t.Errorf("Item %d: Expected CPUs %d, got %d", i, original.Items[i].Spec.CPUs, copy.Items[i].Spec.CPUs)
		}
	}
}

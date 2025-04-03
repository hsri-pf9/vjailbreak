
/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
 	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource/resourcestrategy"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Flavor
// +k8s:openapi-gen=true
type Flavor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlavorSpec   `json:"spec,omitempty"`
	Status FlavorStatus `json:"status,omitempty"`
}

// FlavorList
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type FlavorList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Flavor `json:"items"`
}

// DeepCopyObject implements runtime.Object interface
func (in *Flavor) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopy creates a deep copy of Flavor
func (in *Flavor) DeepCopy() *Flavor {
	if in == nil {
		return nil
	}
	out := new(Flavor)
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return out
}

// DeepCopyObject implements runtime.Object interface
func (in *FlavorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopy creates a deep copy of FlavorList
func (in *FlavorList) DeepCopy() *FlavorList {
	if in == nil {
		return nil
	}
	out := new(FlavorList)
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Flavor, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return out
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *Flavor) DeepCopyInto(out *Flavor) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// FlavorSpec defines the desired state of Flavor
type FlavorSpec struct {
}

var _ resource.Object = &Flavor{}
var _ resourcestrategy.Validater = &Flavor{}

func (in *Flavor) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *Flavor) NamespaceScoped() bool {
	return false
}

func (in *Flavor) New() runtime.Object {
	return &Flavor{}
}

func (in *Flavor) NewList() runtime.Object {
	return &FlavorList{}
}

func (in *Flavor) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "vjailbreak.pf9.io",
		Version:  "v1alpha1",
		Resource: "flavors",
	}
}

func (in *Flavor) IsStorageVersion() bool {
	return true
}

func (in *Flavor) Validate(ctx context.Context) field.ErrorList {
	// TODO(user): Modify it, adding your API validation here.
	return nil
}

var _ resource.ObjectList = &FlavorList{}

func (in *FlavorList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}
// FlavorStatus defines the observed state of Flavor
type FlavorStatus struct {
}

func (in FlavorStatus) SubResourceName() string {
	return "status"
}

// Flavor implements ObjectWithStatusSubResource interface.
var _ resource.ObjectWithStatusSubResource = &Flavor{}

func (in *Flavor) GetStatus() resource.StatusSubResource {
	return in.Status
}

// FlavorStatus{} implements StatusSubResource interface.
var _ resource.StatusSubResource = &FlavorStatus{}

func (in FlavorStatus) CopyTo(parent resource.ObjectWithStatusSubResource) {
	parent.(*Flavor).Status = in
}

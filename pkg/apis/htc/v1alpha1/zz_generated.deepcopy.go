// +build !ignore_autogenerated

// Code generated by operator-sdk. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTCJob) DeepCopyInto(out *HTCJob) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTCJob.
func (in *HTCJob) DeepCopy() *HTCJob {
	if in == nil {
		return nil
	}
	out := new(HTCJob)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HTCJob) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTCJobList) DeepCopyInto(out *HTCJobList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]HTCJob, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTCJobList.
func (in *HTCJobList) DeepCopy() *HTCJobList {
	if in == nil {
		return nil
	}
	out := new(HTCJobList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HTCJobList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTCJobSpec) DeepCopyInto(out *HTCJobSpec) {
	*out = *in
	out.Script = in.Script
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTCJobSpec.
func (in *HTCJobSpec) DeepCopy() *HTCJobSpec {
	if in == nil {
		return nil
	}
	out := new(HTCJobSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTCJobStatus) DeepCopyInto(out *HTCJobStatus) {
	*out = *in
	if in.JobIDs != nil {
		in, out := &in.JobIDs, &out.JobIDs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTCJobStatus.
func (in *HTCJobStatus) DeepCopy() *HTCJobStatus {
	if in == nil {
		return nil
	}
	out := new(HTCJobStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScriptSpec) DeepCopyInto(out *ScriptSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScriptSpec.
func (in *ScriptSpec) DeepCopy() *ScriptSpec {
	if in == nil {
		return nil
	}
	out := new(ScriptSpec)
	in.DeepCopyInto(out)
	return out
}

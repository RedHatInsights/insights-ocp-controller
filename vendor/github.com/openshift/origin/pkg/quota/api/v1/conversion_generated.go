// +build !ignore_autogenerated_openshift

// This file was autogenerated by conversion-gen. Do not edit it manually!

package v1

import (
	quota_api "github.com/openshift/origin/pkg/quota/api"
	api "k8s.io/kubernetes/pkg/api"
	api_v1 "k8s.io/kubernetes/pkg/api/v1"
	conversion "k8s.io/kubernetes/pkg/conversion"
)

func init() {
	if err := api.Scheme.AddGeneratedConversionFuncs(
		Convert_v1_AppliedClusterResourceQuota_To_api_AppliedClusterResourceQuota,
		Convert_api_AppliedClusterResourceQuota_To_v1_AppliedClusterResourceQuota,
		Convert_v1_AppliedClusterResourceQuotaList_To_api_AppliedClusterResourceQuotaList,
		Convert_api_AppliedClusterResourceQuotaList_To_v1_AppliedClusterResourceQuotaList,
		Convert_v1_ClusterResourceQuota_To_api_ClusterResourceQuota,
		Convert_api_ClusterResourceQuota_To_v1_ClusterResourceQuota,
		Convert_v1_ClusterResourceQuotaList_To_api_ClusterResourceQuotaList,
		Convert_api_ClusterResourceQuotaList_To_v1_ClusterResourceQuotaList,
		Convert_v1_ClusterResourceQuotaSelector_To_api_ClusterResourceQuotaSelector,
		Convert_api_ClusterResourceQuotaSelector_To_v1_ClusterResourceQuotaSelector,
		Convert_v1_ClusterResourceQuotaSpec_To_api_ClusterResourceQuotaSpec,
		Convert_api_ClusterResourceQuotaSpec_To_v1_ClusterResourceQuotaSpec,
		Convert_v1_ClusterResourceQuotaStatus_To_api_ClusterResourceQuotaStatus,
		Convert_api_ClusterResourceQuotaStatus_To_v1_ClusterResourceQuotaStatus,
	); err != nil {
		// if one of the conversion functions is malformed, detect it immediately.
		panic(err)
	}
}

func autoConvert_v1_AppliedClusterResourceQuota_To_api_AppliedClusterResourceQuota(in *AppliedClusterResourceQuota, out *quota_api.AppliedClusterResourceQuota, s conversion.Scope) error {
	if err := api.Convert_unversioned_TypeMeta_To_unversioned_TypeMeta(&in.TypeMeta, &out.TypeMeta, s); err != nil {
		return err
	}
	if err := api_v1.Convert_v1_ObjectMeta_To_api_ObjectMeta(&in.ObjectMeta, &out.ObjectMeta, s); err != nil {
		return err
	}
	if err := Convert_v1_ClusterResourceQuotaSpec_To_api_ClusterResourceQuotaSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_v1_ClusterResourceQuotaStatus_To_api_ClusterResourceQuotaStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

func Convert_v1_AppliedClusterResourceQuota_To_api_AppliedClusterResourceQuota(in *AppliedClusterResourceQuota, out *quota_api.AppliedClusterResourceQuota, s conversion.Scope) error {
	return autoConvert_v1_AppliedClusterResourceQuota_To_api_AppliedClusterResourceQuota(in, out, s)
}

func autoConvert_api_AppliedClusterResourceQuota_To_v1_AppliedClusterResourceQuota(in *quota_api.AppliedClusterResourceQuota, out *AppliedClusterResourceQuota, s conversion.Scope) error {
	if err := api.Convert_unversioned_TypeMeta_To_unversioned_TypeMeta(&in.TypeMeta, &out.TypeMeta, s); err != nil {
		return err
	}
	if err := api_v1.Convert_api_ObjectMeta_To_v1_ObjectMeta(&in.ObjectMeta, &out.ObjectMeta, s); err != nil {
		return err
	}
	if err := Convert_api_ClusterResourceQuotaSpec_To_v1_ClusterResourceQuotaSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_api_ClusterResourceQuotaStatus_To_v1_ClusterResourceQuotaStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

func Convert_api_AppliedClusterResourceQuota_To_v1_AppliedClusterResourceQuota(in *quota_api.AppliedClusterResourceQuota, out *AppliedClusterResourceQuota, s conversion.Scope) error {
	return autoConvert_api_AppliedClusterResourceQuota_To_v1_AppliedClusterResourceQuota(in, out, s)
}

func autoConvert_v1_AppliedClusterResourceQuotaList_To_api_AppliedClusterResourceQuotaList(in *AppliedClusterResourceQuotaList, out *quota_api.AppliedClusterResourceQuotaList, s conversion.Scope) error {
	if err := api.Convert_unversioned_TypeMeta_To_unversioned_TypeMeta(&in.TypeMeta, &out.TypeMeta, s); err != nil {
		return err
	}
	if err := api.Convert_unversioned_ListMeta_To_unversioned_ListMeta(&in.ListMeta, &out.ListMeta, s); err != nil {
		return err
	}
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]quota_api.AppliedClusterResourceQuota, len(*in))
		for i := range *in {
			if err := Convert_v1_AppliedClusterResourceQuota_To_api_AppliedClusterResourceQuota(&(*in)[i], &(*out)[i], s); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

func Convert_v1_AppliedClusterResourceQuotaList_To_api_AppliedClusterResourceQuotaList(in *AppliedClusterResourceQuotaList, out *quota_api.AppliedClusterResourceQuotaList, s conversion.Scope) error {
	return autoConvert_v1_AppliedClusterResourceQuotaList_To_api_AppliedClusterResourceQuotaList(in, out, s)
}

func autoConvert_api_AppliedClusterResourceQuotaList_To_v1_AppliedClusterResourceQuotaList(in *quota_api.AppliedClusterResourceQuotaList, out *AppliedClusterResourceQuotaList, s conversion.Scope) error {
	if err := api.Convert_unversioned_TypeMeta_To_unversioned_TypeMeta(&in.TypeMeta, &out.TypeMeta, s); err != nil {
		return err
	}
	if err := api.Convert_unversioned_ListMeta_To_unversioned_ListMeta(&in.ListMeta, &out.ListMeta, s); err != nil {
		return err
	}
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AppliedClusterResourceQuota, len(*in))
		for i := range *in {
			if err := Convert_api_AppliedClusterResourceQuota_To_v1_AppliedClusterResourceQuota(&(*in)[i], &(*out)[i], s); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

func Convert_api_AppliedClusterResourceQuotaList_To_v1_AppliedClusterResourceQuotaList(in *quota_api.AppliedClusterResourceQuotaList, out *AppliedClusterResourceQuotaList, s conversion.Scope) error {
	return autoConvert_api_AppliedClusterResourceQuotaList_To_v1_AppliedClusterResourceQuotaList(in, out, s)
}

func autoConvert_v1_ClusterResourceQuota_To_api_ClusterResourceQuota(in *ClusterResourceQuota, out *quota_api.ClusterResourceQuota, s conversion.Scope) error {
	if err := api.Convert_unversioned_TypeMeta_To_unversioned_TypeMeta(&in.TypeMeta, &out.TypeMeta, s); err != nil {
		return err
	}
	if err := api_v1.Convert_v1_ObjectMeta_To_api_ObjectMeta(&in.ObjectMeta, &out.ObjectMeta, s); err != nil {
		return err
	}
	if err := Convert_v1_ClusterResourceQuotaSpec_To_api_ClusterResourceQuotaSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_v1_ClusterResourceQuotaStatus_To_api_ClusterResourceQuotaStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

func Convert_v1_ClusterResourceQuota_To_api_ClusterResourceQuota(in *ClusterResourceQuota, out *quota_api.ClusterResourceQuota, s conversion.Scope) error {
	return autoConvert_v1_ClusterResourceQuota_To_api_ClusterResourceQuota(in, out, s)
}

func autoConvert_api_ClusterResourceQuota_To_v1_ClusterResourceQuota(in *quota_api.ClusterResourceQuota, out *ClusterResourceQuota, s conversion.Scope) error {
	if err := api.Convert_unversioned_TypeMeta_To_unversioned_TypeMeta(&in.TypeMeta, &out.TypeMeta, s); err != nil {
		return err
	}
	if err := api_v1.Convert_api_ObjectMeta_To_v1_ObjectMeta(&in.ObjectMeta, &out.ObjectMeta, s); err != nil {
		return err
	}
	if err := Convert_api_ClusterResourceQuotaSpec_To_v1_ClusterResourceQuotaSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_api_ClusterResourceQuotaStatus_To_v1_ClusterResourceQuotaStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

func Convert_api_ClusterResourceQuota_To_v1_ClusterResourceQuota(in *quota_api.ClusterResourceQuota, out *ClusterResourceQuota, s conversion.Scope) error {
	return autoConvert_api_ClusterResourceQuota_To_v1_ClusterResourceQuota(in, out, s)
}

func autoConvert_v1_ClusterResourceQuotaList_To_api_ClusterResourceQuotaList(in *ClusterResourceQuotaList, out *quota_api.ClusterResourceQuotaList, s conversion.Scope) error {
	if err := api.Convert_unversioned_TypeMeta_To_unversioned_TypeMeta(&in.TypeMeta, &out.TypeMeta, s); err != nil {
		return err
	}
	if err := api.Convert_unversioned_ListMeta_To_unversioned_ListMeta(&in.ListMeta, &out.ListMeta, s); err != nil {
		return err
	}
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]quota_api.ClusterResourceQuota, len(*in))
		for i := range *in {
			if err := Convert_v1_ClusterResourceQuota_To_api_ClusterResourceQuota(&(*in)[i], &(*out)[i], s); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

func Convert_v1_ClusterResourceQuotaList_To_api_ClusterResourceQuotaList(in *ClusterResourceQuotaList, out *quota_api.ClusterResourceQuotaList, s conversion.Scope) error {
	return autoConvert_v1_ClusterResourceQuotaList_To_api_ClusterResourceQuotaList(in, out, s)
}

func autoConvert_api_ClusterResourceQuotaList_To_v1_ClusterResourceQuotaList(in *quota_api.ClusterResourceQuotaList, out *ClusterResourceQuotaList, s conversion.Scope) error {
	if err := api.Convert_unversioned_TypeMeta_To_unversioned_TypeMeta(&in.TypeMeta, &out.TypeMeta, s); err != nil {
		return err
	}
	if err := api.Convert_unversioned_ListMeta_To_unversioned_ListMeta(&in.ListMeta, &out.ListMeta, s); err != nil {
		return err
	}
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ClusterResourceQuota, len(*in))
		for i := range *in {
			if err := Convert_api_ClusterResourceQuota_To_v1_ClusterResourceQuota(&(*in)[i], &(*out)[i], s); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

func Convert_api_ClusterResourceQuotaList_To_v1_ClusterResourceQuotaList(in *quota_api.ClusterResourceQuotaList, out *ClusterResourceQuotaList, s conversion.Scope) error {
	return autoConvert_api_ClusterResourceQuotaList_To_v1_ClusterResourceQuotaList(in, out, s)
}

func autoConvert_v1_ClusterResourceQuotaSelector_To_api_ClusterResourceQuotaSelector(in *ClusterResourceQuotaSelector, out *quota_api.ClusterResourceQuotaSelector, s conversion.Scope) error {
	out.LabelSelector = in.LabelSelector
	out.AnnotationSelector = in.AnnotationSelector
	return nil
}

func Convert_v1_ClusterResourceQuotaSelector_To_api_ClusterResourceQuotaSelector(in *ClusterResourceQuotaSelector, out *quota_api.ClusterResourceQuotaSelector, s conversion.Scope) error {
	return autoConvert_v1_ClusterResourceQuotaSelector_To_api_ClusterResourceQuotaSelector(in, out, s)
}

func autoConvert_api_ClusterResourceQuotaSelector_To_v1_ClusterResourceQuotaSelector(in *quota_api.ClusterResourceQuotaSelector, out *ClusterResourceQuotaSelector, s conversion.Scope) error {
	out.LabelSelector = in.LabelSelector
	out.AnnotationSelector = in.AnnotationSelector
	return nil
}

func Convert_api_ClusterResourceQuotaSelector_To_v1_ClusterResourceQuotaSelector(in *quota_api.ClusterResourceQuotaSelector, out *ClusterResourceQuotaSelector, s conversion.Scope) error {
	return autoConvert_api_ClusterResourceQuotaSelector_To_v1_ClusterResourceQuotaSelector(in, out, s)
}

func autoConvert_v1_ClusterResourceQuotaSpec_To_api_ClusterResourceQuotaSpec(in *ClusterResourceQuotaSpec, out *quota_api.ClusterResourceQuotaSpec, s conversion.Scope) error {
	if err := Convert_v1_ClusterResourceQuotaSelector_To_api_ClusterResourceQuotaSelector(&in.Selector, &out.Selector, s); err != nil {
		return err
	}
	if err := api_v1.Convert_v1_ResourceQuotaSpec_To_api_ResourceQuotaSpec(&in.Quota, &out.Quota, s); err != nil {
		return err
	}
	return nil
}

func Convert_v1_ClusterResourceQuotaSpec_To_api_ClusterResourceQuotaSpec(in *ClusterResourceQuotaSpec, out *quota_api.ClusterResourceQuotaSpec, s conversion.Scope) error {
	return autoConvert_v1_ClusterResourceQuotaSpec_To_api_ClusterResourceQuotaSpec(in, out, s)
}

func autoConvert_api_ClusterResourceQuotaSpec_To_v1_ClusterResourceQuotaSpec(in *quota_api.ClusterResourceQuotaSpec, out *ClusterResourceQuotaSpec, s conversion.Scope) error {
	if err := Convert_api_ClusterResourceQuotaSelector_To_v1_ClusterResourceQuotaSelector(&in.Selector, &out.Selector, s); err != nil {
		return err
	}
	if err := api_v1.Convert_api_ResourceQuotaSpec_To_v1_ResourceQuotaSpec(&in.Quota, &out.Quota, s); err != nil {
		return err
	}
	return nil
}

func Convert_api_ClusterResourceQuotaSpec_To_v1_ClusterResourceQuotaSpec(in *quota_api.ClusterResourceQuotaSpec, out *ClusterResourceQuotaSpec, s conversion.Scope) error {
	return autoConvert_api_ClusterResourceQuotaSpec_To_v1_ClusterResourceQuotaSpec(in, out, s)
}

func autoConvert_v1_ClusterResourceQuotaStatus_To_api_ClusterResourceQuotaStatus(in *ClusterResourceQuotaStatus, out *quota_api.ClusterResourceQuotaStatus, s conversion.Scope) error {
	if err := api_v1.Convert_v1_ResourceQuotaStatus_To_api_ResourceQuotaStatus(&in.Total, &out.Total, s); err != nil {
		return err
	}
	if err := Convert_v1_ResourceQuotasStatusByNamespace_To_api_ResourceQuotasStatusByNamespace(&in.Namespaces, &out.Namespaces, s); err != nil {
		return err
	}
	return nil
}

func Convert_v1_ClusterResourceQuotaStatus_To_api_ClusterResourceQuotaStatus(in *ClusterResourceQuotaStatus, out *quota_api.ClusterResourceQuotaStatus, s conversion.Scope) error {
	return autoConvert_v1_ClusterResourceQuotaStatus_To_api_ClusterResourceQuotaStatus(in, out, s)
}

func autoConvert_api_ClusterResourceQuotaStatus_To_v1_ClusterResourceQuotaStatus(in *quota_api.ClusterResourceQuotaStatus, out *ClusterResourceQuotaStatus, s conversion.Scope) error {
	if err := api_v1.Convert_api_ResourceQuotaStatus_To_v1_ResourceQuotaStatus(&in.Total, &out.Total, s); err != nil {
		return err
	}
	if err := Convert_api_ResourceQuotasStatusByNamespace_To_v1_ResourceQuotasStatusByNamespace(&in.Namespaces, &out.Namespaces, s); err != nil {
		return err
	}
	return nil
}

func Convert_api_ClusterResourceQuotaStatus_To_v1_ClusterResourceQuotaStatus(in *quota_api.ClusterResourceQuotaStatus, out *ClusterResourceQuotaStatus, s conversion.Scope) error {
	return autoConvert_api_ClusterResourceQuotaStatus_To_v1_ClusterResourceQuotaStatus(in, out, s)
}

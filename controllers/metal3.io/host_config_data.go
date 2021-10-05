package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metal3v1alpha1 "github.com/yashpatil17/baremetal-operator/apis/metal3.io/v1alpha1"
)

// hostConfigData is an implementation of host configuration data interface.
// Object is able to retrive data from secrets referenced in a host spec
type hostConfigData struct {
	host      *metal3v1alpha1.BareMetalHost
	log       logr.Logger
	client    client.Client
	apiReader client.Reader
}

// Generic method for retrieving a secret. If the secret is not found in the
// filtered cache, then the object is retrieved directly from the API
func getSecret(client client.Client, apiReader client.Reader, secretKey types.NamespacedName) (secret *corev1.Secret, err error) {

	secret = &corev1.Secret{}

	// Look for secret in the filtered cache
	err = client.Get(context.TODO(), secretKey, secret)
	if err == nil {
		return secret, nil
	}
	if !k8serrors.IsNotFound(err) {
		return nil, err
	}

	// Secret not in cache; check API directly for unlabelled Secret
	err = apiReader.Get(context.TODO(), secretKey, secret)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, err
		}
		return nil, err
	}

	return secret, nil
}

// Generic method for data extraction from a Secret. Function uses dataKey
// parameter to detirmine which data to return in case secret contins multiple
// keys
func (hcd *hostConfigData) getSecretData(name, namespace, dataKey string) (string, error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	secret, err := getSecret(hcd.client, hcd.apiReader, key)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to fetch user data from secret %s defined in namespace %s", name, namespace))
	}

	if !metav1.HasLabel(secret.ObjectMeta, LabelEnvironmentName) {
		hcd.log.Info("updating secret environment label", "secret", secret.Name, "namespace", secret.Namespace)
		metav1.SetMetaDataLabel(&secret.ObjectMeta, LabelEnvironmentName, LabelEnvironmentValue)

		err = hcd.client.Update(context.TODO(), secret)
		if err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("failed to updated secret %s defined in namespace %s", name, namespace))
		}
	}

	data, ok := secret.Data[dataKey]
	if ok {
		return string(data), nil
	}
	// There is no data under dataKey (userData or networkData).
	// Tring to falback to 'value' key
	if data, ok = secret.Data["value"]; !ok {
		hostConfigDataError.WithLabelValues(dataKey).Inc()
		return "", NoDataInSecretError{secret: name, key: dataKey}
	}

	return string(data), nil
}

// UserData get Operating System configuration data
func (hcd *hostConfigData) UserData() (string, error) {
	if hcd.host.Spec.UserData == nil {
		hcd.log.Info("UserData is not set return empty string")
		return "", nil
	}
	namespace := hcd.host.Spec.UserData.Namespace
	if namespace == "" {
		namespace = hcd.host.Namespace
	}
	return hcd.getSecretData(
		hcd.host.Spec.UserData.Name,
		namespace,
		"userData",
	)

}

// NetworkData get network configuration
func (hcd *hostConfigData) NetworkData() (string, error) {
	if hcd.host.Spec.NetworkData == nil {
		hcd.log.Info("NetworkData is not set returning epmty(nil) data")
		return "", nil
	}
	namespace := hcd.host.Spec.NetworkData.Namespace
	if namespace == "" {
		namespace = hcd.host.Namespace
	}
	return hcd.getSecretData(
		hcd.host.Spec.NetworkData.Name,
		namespace,
		"networkData",
	)
}

// MetaData get host metatdata
func (hcd *hostConfigData) MetaData() (string, error) {
	if hcd.host.Spec.MetaData == nil {
		hcd.log.Info("MetaData is not set returning empty(nil) data")
		return "", nil
	}
	namespace := hcd.host.Spec.MetaData.Namespace
	if namespace == "" {
		namespace = hcd.host.Namespace
	}
	return hcd.getSecretData(
		hcd.host.Spec.MetaData.Name,
		namespace,
		"metaData",
	)
}

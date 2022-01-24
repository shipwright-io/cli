package mock

import (
	"bytes"
	"io/ioutil"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
)

// FakeClientset creates a fake-client to simulate the remote-execution (exec) against a pod. It
// simulates the low level Kubernetes API calls in order to intercept stdin, stdout and stderr sent
// back to the client.
type FakeClientset struct {
	scheme *runtime.Scheme
	codecs serializer.CodecFactory
	codec  runtime.Codec

	restConfig *rest.Config
	clientset  *kubernetes.Clientset

	pod *corev1.Pod
}

// RESTConfig exposes the restConfig attribute.
func (f *FakeClientset) RESTConfig() *rest.Config {
	return f.restConfig
}

// Clientset exposes the clientset attribute.
func (f *FakeClientset) Clientset() *kubernetes.Clientset {
	return f.clientset
}

// roundTripperFn handles the request against the Kubernetes API, and ultimately returns the object
// requested by the client.
func (f *FakeClientset) roundTripperFn(req *http.Request) (*http.Response, error) {
	switch method := req.Method; {
	case method == "GET":
		body := ioutil.NopCloser(bytes.NewReader([]byte(runtime.EncodeOrDie(f.codec, f.pod))))
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
			Header:     map[string][]string{"Content-Type": {"application/json"}},
		}, nil
	default:
		return nil, nil
	}
}

// bootstrap instantiate the basic elements of the clientset.
func (f *FakeClientset) bootstrap() {
	f.scheme = runtime.NewScheme()
	f.scheme.AddIgnoredConversionType(&metav1.TypeMeta{}, &metav1.TypeMeta{})
	f.scheme.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.Pod{}, &metav1.Status{})

	f.codecs = serializer.NewCodecFactory(f.scheme)
	f.codec = f.codecs.LegacyCodec(corev1.SchemeGroupVersion)
}

// instantiateClientset instantiate the clientset using local runtime information and overwriting
// a few resource clients with local instance.  Also, using a fake.RESTClient with local handler
// function.
func (f *FakeClientset) instantiateClientset() {
	httpClient := fake.CreateHTTPClient(f.roundTripperFn)
	restClient := fake.RESTClient{Client: httpClient}
	f.restConfig = &rest.Config{
		ContentConfig: rest.ContentConfig{
			NegotiatedSerializer: f.codecs,
			GroupVersion:         &schema.GroupVersion{Version: corev1.SchemeGroupVersion.Version},
		},
	}
	f.clientset = kubernetes.NewForConfigOrDie(f.restConfig)
	f.clientset.CoreV1().RESTClient().(*rest.RESTClient).Client = restClient.Client
	f.clientset.ExtensionsV1beta1().RESTClient().(*rest.RESTClient).Client = restClient.Client
}

// NewFakeClientset instantiate fake client with informed pod.
func NewFakeClientset(pod *corev1.Pod) *FakeClientset {
	f := &FakeClientset{pod: pod}
	f.bootstrap()
	f.instantiateClientset()
	return f
}

package suites

import (
	"context"
	"flag"
	"testing"

	fluxhelmv2 "github.com/fluxcd/helm-controller/api/v2"
	"github.com/mesosphere/kommander-applications/apptests/docker"
	"github.com/mesosphere/kommander-applications/apptests/environment"
	"github.com/mesosphere/kommander-applications/apptests/flux"
	"github.com/mesosphere/kommander-applications/apptests/kind"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	genericClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	env              *environment.Env
	ctx              context.Context
	network          *docker.NetworkResource
	k8sClient        genericClient.Client
	restClientV1Pods rest.Interface
)

var appVersion = flag.String("app-version", "6.9.4", "The version of the application")

var _ = BeforeSuite(func() {
	// Initialize controller-runtime logger
	log.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx = context.Background()
	var err error
	network, err = kind.EnsureDockerNetworkExist(ctx, "", false)
	Expect(err).ShouldNot(HaveOccurred())

	env = &environment.Env{
		Network: network,
	}
})

func TestApplications(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, reporterConfig := GinkgoConfiguration()
	RunSpecs(t, "Application Test Suite", suiteConfig, reporterConfig)
}

func SetupKindCluster() error {
	if ctx == nil {
		ctx = context.Background()
	}

	err := env.Provision(ctx)
	if err != nil {
		return err
	}

	scheme := flux.NewScheme()
	// Add helm v2 to the scheme because not available in flux.NewScheme()
	_ = fluxhelmv2.AddToScheme(scheme)

	k8sClient, err = genericClient.New(env.K8sClient.Config(), genericClient.Options{Scheme: scheme})
	if err != nil {
		return err
	}

	// Get a REST client for making http requests to pods
	gvk := schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Pod",
	}

	httpClient, err := rest.HTTPClientFor(env.K8sClient.Config())
	if err != nil {
		return err
	}

	restClientV1Pods, err = apiutil.RESTClientForGVK(
		gvk,
		false,
		false,
		env.K8sClient.Config(),
		serializer.NewCodecFactory(flux.NewScheme()),
		httpClient,
	)
	if err != nil {
		return err
	}

	return nil
}

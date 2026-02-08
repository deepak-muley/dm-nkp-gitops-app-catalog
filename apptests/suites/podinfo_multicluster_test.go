package suites

import (
	"context"
	"fmt"
	"os"
	"time"

	fluxhelmv2 "github.com/fluxcd/helm-controller/api/v2"
	apimeta "github.com/fluxcd/pkg/apis/meta"
	"github.com/mesosphere/kommander-applications/apptests/flux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	genericClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/deepak-muley/dm-nkp-gitops-app-catalog/apptests/appscenarios"
	"github.com/deepak-muley/dm-nkp-gitops-app-catalog/apptests/appscenarios/constant"
)

var _ = Describe("Podinfo multicluster", Ordered, Label("podinfo", "multicluster"), func() {
	var (
		workloadClient genericClient.Client
		pr             *appscenarios.PodinfoMulticluster
	)

	BeforeEach(OncePerOrdered, func() {
		if ctx == nil {
			ctx = context.Background()
		}
		err := env.ProvisionMultiCluster(ctx)
		Expect(err).ToNot(HaveOccurred())

		err = env.InstallLatestFlux(ctx)
		Expect(err).ToNot(HaveOccurred())

		err = env.InstallLatestFluxOnWorkload(ctx)
		Expect(err).ToNot(HaveOccurred())

		scheme := flux.NewScheme()
		_ = fluxhelmv2.AddToScheme(scheme)
		workloadClient, err = genericClient.New(env.WorkloadK8sClient.Config(), genericClient.Options{Scheme: scheme})
		Expect(err).ToNot(HaveOccurred())

		pr = &appscenarios.PodinfoMulticluster{VersionToInstall: ""}
	})

	AfterEach(OncePerOrdered, func() {
		if os.Getenv("SKIP_CLUSTER_TEARDOWN") != "" {
			return
		}
		err := env.DestroyMultiCluster(ctx)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should install podinfo on workload cluster", func() {
		err := pr.InstallToWorkload(ctx, env)
		Expect(err).ToNot(HaveOccurred())

		hr := &fluxhelmv2.HelmRelease{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pr.Name(),
				Namespace: constant.DEFAULT_NAMESPACE,
			},
		}
		Eventually(func() error {
			err := workloadClient.Get(ctx, ctrlClient.ObjectKeyFromObject(hr), hr)
			if err != nil {
				return err
			}
			for _, cond := range hr.Status.Conditions {
				if cond.Status == metav1.ConditionTrue && cond.Type == apimeta.ReadyCondition {
					return nil
				}
			}
			return fmt.Errorf("helm release not ready yet")
		}).WithPolling(constant.POLL_INTERVAL).WithTimeout(5 * time.Minute).Should(Succeed())
	})
})

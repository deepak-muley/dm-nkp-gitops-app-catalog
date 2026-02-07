package suites

import (
	"fmt"
	"os"
	"time"

	fluxhelmv2 "github.com/fluxcd/helm-controller/api/v2"
	apimeta "github.com/fluxcd/pkg/apis/meta"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/deepak-muley/dm-nkp-gitops-app-catalog/apptests/appscenarios"
	"github.com/deepak-muley/dm-nkp-gitops-app-catalog/apptests/appscenarios/constant"
)

var _ = Describe("podinfo Tests", Ordered, Label("podinfo"), func() {
	// Print values passed from CLI flags
	BeforeEach(OncePerOrdered, func() {
		err := SetupKindCluster()
		Expect(err).ToNot(HaveOccurred())

		err = env.InstallLatestFlux(ctx)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(OncePerOrdered, func() {
		if os.Getenv("SKIP_CLUSTER_TEARDOWN") != "" {
			return
		}

		err := env.Destroy(ctx)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Installing podinfo", Ordered, Label("install"), func() {
		var (
			pr *appscenarios.Podinfo
			hr *fluxhelmv2.HelmRelease
		)

		It("should install successfully with default config", func() {
			pr = appscenarios.NewPodinfoScenerio(*appVersion).(*appscenarios.Podinfo)
			err := pr.Install(ctx, env)
			Expect(err).ToNot(HaveOccurred())

			hr = &fluxhelmv2.HelmRelease{
				TypeMeta: metav1.TypeMeta{
					Kind:       fluxhelmv2.HelmReleaseKind,
					APIVersion: fluxhelmv2.GroupVersion.Version,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      pr.Name(),
					Namespace: constant.DEFAULT_NAMESPACE,
				},
			}

			// Check the status of the HelmReleases
			Eventually(func() error {
				err = k8sClient.Get(ctx, ctrlClient.ObjectKeyFromObject(hr), hr)
				if err != nil {
					return err
				}

				for _, cond := range hr.Status.Conditions {
					if cond.Status == metav1.ConditionTrue &&
						cond.Type == apimeta.ReadyCondition {
						return nil
					}
				}
				return fmt.Errorf("helm release not ready yet")
			}).WithPolling(constant.POLL_INTERVAL).WithTimeout(5 * time.Minute).Should(Succeed())
		})
	})

	Describe("Upgrading podinfo", Ordered, Label("upgrade"), func() {
		var (
			pr *appscenarios.Podinfo
			hr *fluxhelmv2.HelmRelease
		)

		It("should install the previous version successfully", func() {
			pr = appscenarios.NewPodinfoScenerio("").(*appscenarios.Podinfo)
			err := pr.InstallPreviousVersion(ctx, env)
			Expect(err).ToNot(HaveOccurred())

			hr = &fluxhelmv2.HelmRelease{
				TypeMeta: metav1.TypeMeta{
					Kind:       fluxhelmv2.HelmReleaseKind,
					APIVersion: fluxhelmv2.GroupVersion.Version,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      pr.Name(),
					Namespace: constant.DEFAULT_NAMESPACE,
				},
			}

			// Check the status of the HelmReleases
			Eventually(func() error {
				err = k8sClient.Get(ctx, ctrlClient.ObjectKeyFromObject(hr), hr)
				if err != nil {
					return err
				}

				for _, cond := range hr.Status.Conditions {
					if cond.Status == metav1.ConditionTrue &&
						cond.Type == apimeta.ReadyCondition {
						return nil
					}
				}
				return fmt.Errorf("helm release not ready yet")
			}).WithPolling(constant.POLL_INTERVAL).WithTimeout(5 * time.Minute).Should(Succeed())
		})

		It("should upgrade podinfo successfully", func() {
			err := pr.Upgrade(ctx, env)
			Expect(err).ToNot(HaveOccurred())

			hr = &fluxhelmv2.HelmRelease{
				TypeMeta: metav1.TypeMeta{
					Kind:       fluxhelmv2.HelmReleaseKind,
					APIVersion: fluxhelmv2.GroupVersion.Version,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      pr.Name(),
					Namespace: constant.DEFAULT_NAMESPACE,
				},
			}

			// Check the status of the HelmReleases
			Eventually(func() error {
				err = k8sClient.Get(ctx, ctrlClient.ObjectKeyFromObject(hr), hr)
				if err != nil {
					return err
				}

				for _, cond := range hr.Status.Conditions {
					if cond.Status == metav1.ConditionTrue &&
						cond.Type == apimeta.ReadyCondition &&
						cond.Reason == fluxhelmv2.UpgradeSucceededReason {
						return nil
					}
				}
				return fmt.Errorf("helm release not ready yet")
			}).WithPolling(constant.POLL_INTERVAL).WithTimeout(5 * time.Minute).Should(Succeed())
		})
	})
})

package catalogapptests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/deepak-muley/dm-nkp-gitops-app-catalog/catalog-apptests/framework"
	fluxhelmv2 "github.com/fluxcd/helm-controller/api/v2"
	apimeta "github.com/fluxcd/pkg/apis/meta"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	suiteCtx     context.Context
	suiteNetwork *framework.Network
)

var _ = BeforeSuite(func() {
	log.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	suiteCtx = context.Background()
	var err error
	suiteNetwork, err = framework.EnsureDockerNetworkExist(suiteCtx, "", false)
	Expect(err).ShouldNot(HaveOccurred())
})

func TestCatalogApplications(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, reporterConfig := GinkgoConfiguration()
	RunSpecs(t, "Catalog Application Test Suite", suiteConfig, reporterConfig)
}

func assertHelmReleaseReady(cluster Cluster, name, namespace string, expectUpgradeReason bool) {
	hr := &fluxhelmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	}
	Eventually(func() error {
		err := cluster.Client().Get(cluster.Ctx(), ctrlClient.ObjectKeyFromObject(hr), hr)
		if err != nil {
			return err
		}
		for _, cond := range hr.Status.Conditions {
			if cond.Status == metav1.ConditionTrue && cond.Type == apimeta.ReadyCondition {
				if expectUpgradeReason && cond.Reason != fluxhelmv2.UpgradeSucceededReason {
					return fmt.Errorf("helm release ready but reason=%s", cond.Reason)
				}
				return nil
			}
		}
		return fmt.Errorf("helm release not ready yet")
	}).WithPolling(PollInterval).WithTimeout(5 * time.Minute).Should(Succeed())
}

var _ = Describe("Catalog applications (install/upgrade)", Ordered, Label("templated"), func() {
	catalog, err := DefaultCatalog()
	if err != nil {
		Fail("discovery failed: " + err.Error())
	}
	apps, err := catalog.Apps()
	if err != nil {
		Fail("discovery failed: " + err.Error())
	}

	for i := range apps {
		app := apps[i]
		Describe(app.Name+" install/upgrade", Ordered, Label("appname", app.Name), func() {
			var cluster Cluster

			BeforeEach(OncePerOrdered, func() {
				var err error
				cluster, err = KindCluster.Create(suiteCtx, ClusterConfig{Network: suiteNetwork, Catalog: catalog, Name: "default"})
				Expect(err).ToNot(HaveOccurred())
				Expect(cluster.Install(FluxApp)).ToNot(HaveOccurred())
			})
			AfterEach(OncePerOrdered, func() {
				if os.Getenv("SKIP_CLUSTER_TEARDOWN") != "" {
					return
				}
				cluster.Destroy()
			})

			Describe("Installing "+app.Name, Ordered, Label("install"), func() {
				It("should install successfully with default config", func() {
					catalogApp := NewCatalogApp(app.Name, "")
					Expect(cluster.Install(catalogApp)).ToNot(HaveOccurred())
					assertHelmReleaseReady(cluster, catalogApp.Name(), DefaultNamespace, false)
				})
			})

			if len(app.Versions) >= 2 {
				Describe("Upgrading "+app.Name, Ordered, Label("upgrade"), func() {
					var cat *CatalogApp
					It("should install the previous version successfully", func() {
						cat = NewCatalogApp(app.Name, "")
						Expect(cat.InstallPreviousVersion(cluster)).ToNot(HaveOccurred())
						assertHelmReleaseReady(cluster, cat.Name(), DefaultNamespace, false)
					})
					It("should upgrade successfully", func() {
						if cat == nil {
							cat = NewCatalogApp(app.Name, "")
						}
						Expect(cat.Upgrade(cluster)).ToNot(HaveOccurred())
						assertHelmReleaseReady(cluster, cat.Name(), DefaultNamespace, true)
					})
				})
			}
		})
	}
})

var _ = Describe("Catalog applications (multicluster — OpenCost)", Ordered, Label("templated", "multicluster"), func() {
	catalog, err := DefaultCatalog()
	if err != nil {
		Fail("discovery failed: " + err.Error())
	}
	apps, err := catalog.Apps()
	if err != nil {
		Fail("discovery failed: " + err.Error())
	}

	var centralApp, opencostApp *AppVersions
	for i := range apps {
		if apps[i].Name == MulticlusterCentralAppName {
			centralApp = &apps[i]
		}
		if apps[i].Name == MulticlusterTestAppName {
			opencostApp = &apps[i]
		}
	}

	var mgmt NKPManagementCluster
	var workload1, workload2 NKPWorkloadCluster

	BeforeEach(OncePerOrdered, func() {
		var err error
		var c Cluster
		c, err = KindCluster.Create(suiteCtx, ClusterConfig{Network: suiteNetwork, Catalog: catalog, Name: "mgmt"})
		Expect(err).ToNot(HaveOccurred())
		mgmt = c.(NKPManagementCluster)
		workload1, err = KindCluster.CreateFromParent(suiteCtx, mgmt, "workload1")
		Expect(err).ToNot(HaveOccurred())
		workload2, err = KindCluster.CreateFromParent(suiteCtx, mgmt, "workload2")
		Expect(err).ToNot(HaveOccurred())
		Expect(mgmt.Install(FluxApp)).ToNot(HaveOccurred())
		Expect(workload1.Install(FluxApp)).ToNot(HaveOccurred())
		Expect(workload2.Install(FluxApp)).ToNot(HaveOccurred())
	})
	AfterEach(OncePerOrdered, func() {
		if os.Getenv("SKIP_CLUSTER_TEARDOWN") != "" {
			return
		}
		mgmt.Destroy()
	})

	Describe("OpenCost central on mgmt, clients on workload", Ordered, Label("appname", MulticlusterTestAppName), func() {
		It("should install centralized-opencost on mgmt (central)", func() {
			if centralApp == nil {
				Skip("centralized-opencost not in catalog — add applications/centralized-opencost")
			}
			Expect(mgmt.InstallCentralizedOpencost()).ToNot(HaveOccurred())
			assertHelmReleaseReady(mgmt, "centralized-opencost", DefaultNamespace, false)
		})
		It("should install opencost on workload1 (client)", func() {
			if opencostApp == nil {
				Skip("opencost not in catalog — add applications/opencost")
			}
			Expect(workload1.InstallOpencost()).ToNot(HaveOccurred())
			assertHelmReleaseReady(workload1, "opencost", DefaultNamespace, false)
		})
		It("should install opencost on workload2 (client)", func() {
			if opencostApp == nil {
				Skip("opencost not in catalog — add applications/opencost")
			}
			Expect(workload2.InstallOpencost()).ToNot(HaveOccurred())
			assertHelmReleaseReady(workload2, "opencost", DefaultNamespace, false)
		})
	})
})

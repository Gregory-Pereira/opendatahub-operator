/*
Copyright 2023.

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

package main

import (
	"fmt"
	"os"
	"strings"

	ocappsv1 "github.com/openshift/api/apps/v1" //nolint:importas //reason: conflicts with appsv1 "k8s.io/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	consolev1 "github.com/openshift/api/console/v1"
	imagev1 "github.com/openshift/api/image/v1"
	oauthv1 "github.com/openshift/api/oauth/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	routev1 "github.com/openshift/api/route/v1"
	securityv1 "github.com/openshift/api/security/v1"
	templatev1 "github.com/openshift/api/template/v1"
	userv1 "github.com/openshift/api/user/v1"
	ofapiv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	ofapiv2 "github.com/operator-framework/api/pkg/operators/v2"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	componentApi "github.com/opendatahub-io/opendatahub-operator/v2/api/components/v1alpha1"
	dscv1 "github.com/opendatahub-io/opendatahub-operator/v2/api/datasciencecluster/v1"
	dscv2 "github.com/opendatahub-io/opendatahub-operator/v2/api/datasciencecluster/v2"
	dsciv1 "github.com/opendatahub-io/opendatahub-operator/v2/api/dscinitialization/v1"
	dsciv2 "github.com/opendatahub-io/opendatahub-operator/v2/api/dscinitialization/v2"
	featurev1 "github.com/opendatahub-io/opendatahub-operator/v2/api/features/v1"
	infrav1 "github.com/opendatahub-io/opendatahub-operator/v2/api/infrastructure/v1"
	infrav1alpha1 "github.com/opendatahub-io/opendatahub-operator/v2/api/infrastructure/v1alpha1"
	serviceApi "github.com/opendatahub-io/opendatahub-operator/v2/api/services/v1alpha1"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/logger"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/platform/factory"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/utils/flags"

	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/dashboard"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/datasciencepipelines"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/feastoperator"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/kserve"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/kueue"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/llamastackoperator"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/modelcontroller"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/modelregistry"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/ray"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/trainingoperator"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/trustyai"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/workbenches"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/auth"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/certconfigmapgenerator"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/gateway"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/monitoring"
	_ "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/services/setup"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() { //nolint:gochecknoinits
	utilruntime.Must(componentApi.AddToScheme(scheme))
	utilruntime.Must(serviceApi.AddToScheme(scheme))
	utilruntime.Must(infrav1alpha1.AddToScheme(scheme))
	utilruntime.Must(infrav1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(dsciv1.AddToScheme(scheme))
	utilruntime.Must(dsciv2.AddToScheme(scheme))
	utilruntime.Must(dscv1.AddToScheme(scheme))
	utilruntime.Must(dscv2.AddToScheme(scheme))
	utilruntime.Must(featurev1.AddToScheme(scheme))
	utilruntime.Must(networkingv1.AddToScheme(scheme))
	utilruntime.Must(rbacv1.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(routev1.Install(scheme))
	utilruntime.Must(appsv1.AddToScheme(scheme))
	utilruntime.Must(oauthv1.Install(scheme))
	utilruntime.Must(ofapiv1alpha1.AddToScheme(scheme))
	utilruntime.Must(userv1.Install(scheme))
	utilruntime.Must(ofapiv2.AddToScheme(scheme))
	utilruntime.Must(ocappsv1.Install(scheme))
	utilruntime.Must(buildv1.Install(scheme))
	utilruntime.Must(imagev1.Install(scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
	utilruntime.Must(admissionregistrationv1.AddToScheme(scheme))
	utilruntime.Must(promv1.AddToScheme(scheme))
	utilruntime.Must(operatorv1.Install(scheme))
	utilruntime.Must(consolev1.AddToScheme(scheme))
	utilruntime.Must(securityv1.Install(scheme))
	utilruntime.Must(templatev1.Install(scheme))
	utilruntime.Must(gwapiv1.Install(scheme))
}

func main() { //nolint:funlen
	// Viper settings
	viper.SetEnvPrefix("ODH_MANAGER")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// define flags and env vars
	if err := flags.AddOperatorFlagsAndEnvvars(viper.GetEnvPrefix()); err != nil {
		fmt.Printf("Error in adding flags or binding env vars: %s", err.Error())
		os.Exit(1)
	}

	// parse and bind flags
	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		fmt.Printf("Error in binding flags: %s", err.Error())
		os.Exit(1)
	}

	oconfig, err := cluster.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %s", err.Error())
		os.Exit(1)
	}

	// After getting the zap related configs an ad hoc flag set is created so the zap BindFlags mechanism can be reused
	zapFlagSet := flags.NewZapFlagSet()

	opts := zap.Options{}
	opts.BindFlags(zapFlagSet)

	err = flags.ParseZapFlags(zapFlagSet, oconfig.ZapDevel, oconfig.ZapEncoder, oconfig.ZapLogLevel, oconfig.ZapStacktrace, oconfig.ZapTimeEncoding)
	if err != nil {
		fmt.Printf("Error in parsing zap flags: %s", err.Error())
		os.Exit(1)
	}

	ctrl.SetLogger(logger.NewLogger(oconfig.LogMode, &opts))

	// root context
	ctx := ctrl.SetupSignalHandler()
	ctx = logf.IntoContext(ctx, setupLog)

	// Get rest.Config and attach to OperatorConfig
	setupCfg, err := config.GetConfig()
	if err != nil {
		setupLog.Error(err, "error getting config for setup")
		os.Exit(1)
	}
	oconfig.RestConfig = setupCfg

	// Create new uncached client to run initial setup
	setupClient, err := client.New(setupCfg, client.Options{Scheme: scheme})
	if err != nil {
		setupLog.Error(err, "error getting client for setup")
		os.Exit(1)
	}

	err = cluster.Init(ctx, setupClient)
	if err != nil {
		setupLog.Error(err, "unable to initialize cluster config")
		os.Exit(1)
	}

	// Get operator platform
	release := cluster.GetRelease()

	// Create platform instance
	plat, err := factory.New(release.Name, scheme, oconfig)
	if err != nil {
		setupLog.Error(err, "unable to create platform")
		os.Exit(1)
	}
	setupLog.Info("Platform initialized", "type", plat.String())

	// Initialize platform (handles services and components initialization)
	if err := plat.Init(ctx); err != nil {
		setupLog.Error(err, "unable to initialize platform")
		os.Exit(1)
	}

	// Perform platform-specific upgrade operations
	if err := plat.Upgrade(ctx); err != nil {
		setupLog.Error(err, "unable to perform platform upgrade")
		os.Exit(1)
	}

	// Run platform (creates manager, registers webhooks, starts reconcilers)
	setupLog.Info("Starting platform runtime", "platform", plat.String())
	if err := plat.Run(ctx); err != nil {
		setupLog.Error(err, "problem running platform")
		os.Exit(1)
	}
}

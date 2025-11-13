package cluster

import "github.com/opendatahub-io/opendatahub-operator/v2/api/common"

const (
	// ManagedRhoai defines expected addon catalogsource.
	ManagedRhoai common.Platform = "OpenShift AI Cloud Service"
	// SelfManagedRhoai defines display name in csv.
	SelfManagedRhoai common.Platform = "OpenShift AI Self-Managed"
	// OpenDataHub defines display name in csv.
	OpenDataHub common.Platform = "Open Data Hub"
	// Vanilla defines display name for vanilla Kubernetes deployments.
	Vanilla common.Platform = "Vanilla Kubernetes"

	// DefaultNotebooksNamespaceODH defines default namespace for notebooks.
	DefaultNotebooksNamespaceODH = "opendatahub"
	// DefaultNotebooksNamespaceRHOAI defines default namespace for notebooks.
	DefaultNotebooksNamespaceRHOAI = "rhods-notebooks"

	// DefaultMonitoringNamespaceODH defines default namespace for monitoring.
	DefaultMonitoringNamespaceODH = "opendatahub"
	// DefaultMonitoringNamespaceRHOAI defines default namespace for monitoring.
	DefaultMonitoringNamespaceRHOAI = "redhat-ods-monitoring"

	// DefaultApplicationNamespaceODH defines default namespace for ODH applications.
	DefaultApplicationNamespaceODH = "opendatahub"
	// DefaultApplicationNamespaceRHOAI defines default namespace for RHOAI applications.
	DefaultApplicationNamespaceRHOAI = "redhat-ods-applications"

	// DefaultAdminGroupManaged defines default admin group for managed RHOAI.
	DefaultAdminGroupManaged = "dedicated-admins"
	// DefaultAdminGroupSelfManaged defines default admin group for self-managed RHOAI.
	DefaultAdminGroupSelfManaged = "rhods-admins"
	// DefaultAdminGroupODH defines default admin group for ODH.
	DefaultAdminGroupODH = "odh-admins"

	// DefaultConsoleLinkNamespace defines namespace for console link in managed RHOAI.
	DefaultConsoleLinkNamespace = "redhat-ods-applications-console-link"

	// SubscriptionNameRHOAI defines operator subscription name for RHOAI.
	SubscriptionNameRHOAI = "rhods-operator"
	// SubscriptionNameODH defines operator subscription name for ODH.
	SubscriptionNameODH = "opendatahub-operator"

	// Default cluster-scope Authentication CR name.
	ClusterAuthenticationObj = "cluster"

	// Default OpenShift version CR name.
	OpenShiftVersionObj = "version"

	// Managed cluster required route.
	NameConsoleLink      = "console"
	NamespaceConsoleLink = "openshift-console"

	// KueueQueueNameLabel is the label key used to specify the Kueue queue name for workloads.
	KueueQueueNameLabel = "kueue.x-k8s.io/queue-name"

	// KueueManagedLabelKey indicates a namespace is managed by Kueue.
	KueueManagedLabelKey = "kueue.openshift.io/managed"

	// KueueLegacyManagedLabelKey is the legacy label key used to indicate a namespace is managed by Kueue.
	KueueLegacyManagedLabelKey = "kueue-managed"

	// OpenshiftOperatorsNamespace is the namespace for OpenShift operator dependencies.
	OpenshiftOperatorsNamespace = "openshift-operators"

	// OpenshiftIngressNamespace is the namespace for OpenShift ingress resources.
	OpenshiftIngressNamespace = "openshift-ingress"
)

// api/v1/types.go
// +groupName=quantum.tiktok

package v1

import (
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Sentinel errors returned by the reconciler.
// These are exported so callers can use errors.Is() without
// string matching, which is how we know this codebase has principles.
var (
	ErrQuantumTunneling    = errors.New("dopamine exceeded cringe threshold: tunneling initiated")
	ErrOracleTimeout       = errors.New("oracle did not respond within deadline")
	ErrDecoherenceDetected = errors.New("chaos mesh has destabilized the wavefunction")
	ErrRealityCollapsed    = errors.New("wavefunction collapsed to suboptimal eigenstate")
	ErrMaxDopeExceeded     = errors.New("socialDopamine must not exceed 9000") // we learned this the hard way
)

// CoherenceState represents the observed quantum state of the system.
// +kubebuilder:validation:Enum=coherent;decoherent;superposition
type CoherenceState string

const (
	CoherenceStateCoherent     CoherenceState = "coherent"
	CoherenceStateDecoherent   CoherenceState = "decoherent"
	CoherenceStateSuperposition CoherenceState = "superposition"
)

// OracleResult is the measurement outcome returned by the external oracle.
// +kubebuilder:validation:Enum=✅;💩
type OracleResult string

const (
	OracleResultOptimal    OracleResult = "✅"
	OracleResultSuboptimal OracleResult = "💩"
)

// AnnealingConfig controls the social annealing schedule.
// Temperature decreases over time, reducing the probability of accepting
// worse solutions. At T=0 the system is frozen.
// At T=1000 the system is indistinguishable from a Friday afternoon standup.
type AnnealingConfig struct {
	// InitialTemperature is the starting temperature of the annealing schedule.
	// High values allow the system to escape local minima.
	// Very high values allow the system to escape reality.
	// +kubebuilder:default="1000.0"
	// +kubebuilder:validation:Pattern=`^[0-9]+(\.[0-9]+)?$`
	InitialTemperature string `json:"initialTemperature,omitempty"`

	// CoolingRate is the multiplicative factor applied each iteration.
	// Must be in (0, 1). Values close to 1 cool slowly and find better solutions.
	// Values close to 0 cool instantly and find whatever is nearby.
	// +kubebuilder:default="0.95"
	CoolingRate string `json:"coolingRate,omitempty"`

	// MinTemperature is the lower bound. Below this, annealing stops.
	// +kubebuilder:default="0.01"
	MinTemperature string `json:"minTemperature,omitempty"`
}

// QuantumHamiltonianSpec defines the desired state.
type QuantumHamiltonianSpec struct {
	// SocialDopamine encodes the engagement level of the target workload.
	// Below CringeThreshold: system is stable. Above: tunneling is initiated.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=9000
	SocialDopamine int32 `json:"socialDopamine"`

	// CringeThreshold is the critical value above which the system performs
	// forced pod relocation. Analogous to activation energy in chemistry,
	// but less peer-reviewed.
	// +kubebuilder:validation:Minimum=0
	CringeThreshold int32 `json:"cringeThreshold"`

	// TargetNamespace scopes the operator's influence.
	// Pods outside this namespace are not tunneled. Probably.
	// +kubebuilder:validation:MinLength=1
	TargetNamespace string `json:"targetNamespace"`

	// OracleEndpoint is the URL of the measurement webhook.
	// Must return OracleResultOptimal or OracleResultSuboptimal.
	// All other responses are treated as OracleResultSuboptimal.
	// +kubebuilder:validation:Format=uri
	OracleEndpoint string `json:"oracleEndpoint"`

	// ChaosEnabled controls whether the operator injects decoherence
	// on suboptimal oracle responses. Defaults to true.
	// Disable only with written approval from someone with budget authority.
	// +kubebuilder:default=true
	ChaosEnabled bool `json:"chaosEnabled,omitempty"`

	// Annealing configures the social annealing schedule.
	// +optional
	Annealing AnnealingConfig `json:"annealing,omitempty"`
}

// QuantumHamiltonianStatus reflects the last observed state.
// Fields here are informational only; the reconciler does not act on them.
// The reconciler acts on Spec. Status is for humans and Grafana dashboards
// that will be looked at exactly once, during the incident that caused them to be built.
type QuantumHamiltonianStatus struct {
	// CoherenceState is the current quantum state of the system.
	CoherenceState CoherenceState `json:"coherenceState,omitempty"`

	// LastMeasurement is the most recent oracle response.
	LastMeasurement OracleResult `json:"lastMeasurement,omitempty"`

	// TunnelingCount is the cumulative number of forced pod relocations.
	// High values are normal. Very high values suggest the oracle is broken.
	// Extremely high values suggest the oracle is a meme API.
	TunnelingCount int64 `json:"tunnelingCount,omitempty"`

	// CurrentTemperature is the current annealing temperature.
	// Decreases over time. Cannot increase. Like technical debt.
	CurrentTemperature float64 `json:"currentTemperature,omitempty"`

	// ObservedGeneration is the generation of the Spec last reconciled.
	// Used to detect stale status.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions follows the standard Kubernetes condition pattern.
	// Included because enterprise software has conditions.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// LastTransitionTime records when CoherenceState last changed.
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=qh;qubit,categories=quantum
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.coherenceState`
// +kubebuilder:printcolumn:name="Dopamine",type=integer,JSONPath=`.spec.socialDopamine`
// +kubebuilder:printcolumn:name="Coherence",type=number,JSONPath=`.status.currentTemperature`,format=float
// +kubebuilder:printcolumn:name="Tunneling",type=integer,JSONPath=`.status.tunnelingCount`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type QuantumHamiltonian struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QuantumHamiltonianSpec   `json:"spec,omitempty"`
	Status QuantumHamiltonianStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type QuantumHamiltonianList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QuantumHamiltonian `json:"items"`
}

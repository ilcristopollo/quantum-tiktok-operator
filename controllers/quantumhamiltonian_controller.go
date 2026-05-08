// controllers/quantumhamiltonian_controller.go

package controllers

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	quantumv1 "github.com/yourusername/quantum-tiktok-operator/api/v1"
	"github.com/yourusername/quantum-tiktok-operator/internal/annealing"
	"github.com/yourusername/quantum-tiktok-operator/internal/chaos"
	"github.com/yourusername/quantum-tiktok-operator/internal/oracle"
	"github.com/yourusername/quantum-tiktok-operator/pkg/entanglement"
)

const (
	conditionTypeCoherent    = "Coherent"
	conditionTypeEntangled   = "Entangled"
	requeueOnDecoherence     = 3 * time.Second
	requeueOnCoherence       = 1 * time.Hour
	reconcilerContextTimeout = 2 * time.Second
)

// QuantumHamiltonianReconciler reconciles QuantumHamiltonian objects.
//
// +kubebuilder:rbac:groups=quantum.tiktok,resources=quantumhamiltonians,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=quantum.tiktok,resources=quantumhamiltonians/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=quantum.tiktok,resources=quantumhamiltonians/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;delete
// +kubebuilder:rbac:groups=chaos-mesh.org,resources=networkchaos,verbs=create;delete
type QuantumHamiltonianReconciler struct {
	client.Client
	Oracle      oracle.Client
	Chaos       chaos.Client
	Entanglement entanglement.Registry
	Annealer    *annealing.SocialAnnealer
}

// Reconcile drives the cluster toward the ground state of the Hamiltonian.
// Called on every relevant event. Idempotent. Mostly.
func (r *QuantumHamiltonianReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	ctx, cancel := context.WithTimeout(ctx, reconcilerContextTimeout)
	defer cancel()

	var h quantumv1.QuantumHamiltonian
	if err := r.Get(ctx, req.NamespacedName, &h); err != nil {
		if apierrors.IsNotFound(err) {
			// Resource deleted. Nothing to reconcile.
			// The wavefunction has been manually collapsed by kubectl delete.
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("fetching QuantumHamiltonian: %w", err)
	}

	// Mark in-progress. Prevents stale status from misleading the dashboard
	// during the reconciliation window.
	// The dashboard will be misleading anyway. This just makes it intentional.
	if err := r.setCoherenceState(ctx, &h, quantumv1.CoherenceStateSuperposition); err != nil {
		logger.Error(err, "failed to update status to superposition")
		// Non-fatal: continue reconciliation.
	}

	result, err := r.collapseWaveFunction(ctx, &h)
	if err != nil {
		logger.Error(err, "reconciliation error", "hamiltonian", req.NamespacedName)
		_ = r.setCoherenceState(ctx, &h, quantumv1.CoherenceStateDecoherent)
		return ctrl.Result{RequeueAfter: jitter(requeueOnDecoherence)}, nil
	}

	return result, nil
}

// collapseWaveFunction is the core reconciliation function.
// It drives the system from superposition toward a definite eigenstate.
// The eigenstate is either ✅ or 💩. There is no third option.
func (r *QuantumHamiltonianReconciler) collapseWaveFunction(
	ctx context.Context,
	h *quantumv1.QuantumHamiltonian,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Establish entanglement between managed pods.
	// Entanglement is maintained in etcd. Correlation holds until both
	// endpoints observe the shared state, at which point it collapses.
	//
	// eventual consistency IS the consistency model.
	// stop asking questions.
	if err := r.Entanglement.Entangle(ctx, h.Spec.TargetNamespace, h.Name); err != nil {
		logger.Info("entanglement failed, proceeding without correlation", "err", err)
		// Entanglement failure is non-fatal. The system degrades to classical scheduling,
		// which is slower but less likely to appear in a postmortem.
	}

	// Apply social annealing.
	// If dopamine exceeds the cringe threshold, the system initiates tunneling:
	// a pod is deleted and recreated, passing through the energy barrier
	// of the Kubernetes scheduler.
	if h.Spec.SocialDopamine > h.Spec.CringeThreshold {
		logger.Info("dopamine threshold exceeded, initiating quantum tunneling",
			"dopamine", h.Spec.SocialDopamine,
			"threshold", h.Spec.CringeThreshold,
		)

		if err := r.performQuantumTunneling(ctx, h); err != nil {
			return ctrl.Result{}, fmt.Errorf("%w: %v", quantumv1.ErrQuantumTunneling, err)
		}

		h.Status.TunnelingCount++
	}

	// Update annealing temperature.
	// Temperature decreases each reconciliation cycle.
	// At T ≈ 0 the system no longer accepts worse solutions.
	// At T ≈ 0 on a Friday evening, neither does the on-call engineer.
	newTemp := r.Annealer.Cool(h.Status.CurrentTemperature)
	h.Status.CurrentTemperature = newTemp

	// Measure the current state via the oracle.
	// The oracle is the source of truth. We do not question the oracle.
	// We do, however, set a 5-second timeout on it.
	measurement, err := r.Oracle.Measure(ctx, h.Spec.OracleEndpoint)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("%w: %v", quantumv1.ErrOracleTimeout, err)
	}

	h.Status.LastMeasurement = measurement

	switch measurement {
	case quantumv1.OracleResultOptimal:
		logger.Info("oracle: ✅ — coherent eigenstate reached")
		_ = r.setCoherenceState(ctx, h, quantumv1.CoherenceStateCoherent)
		return ctrl.Result{RequeueAfter: requeueOnCoherence}, nil

	case quantumv1.OracleResultSuboptimal:
		logger.Info("oracle: 💩 — suboptimal eigenstate, injecting decoherence")

		if h.Spec.ChaosEnabled {
			go r.injectDecoherence(h.Spec.TargetNamespace) //nolint:errcheck
		}

		_ = r.setCoherenceState(ctx, h, quantumv1.CoherenceStateDecoherent)
		return ctrl.Result{RequeueAfter: jitter(requeueOnDecoherence)}, nil

	default:
		// The oracle returned something that is neither ✅ nor 💩.
		// This should not be possible given the webhook contract.
		// It is, however, possible given human nature.
		return ctrl.Result{}, fmt.Errorf("unexpected oracle response: %q", measurement)
	}
}

// performQuantumTunneling relocates a pod by deleting it and allowing
// the ReplicaSet controller to recreate it on a different node.
// This is semantically equivalent to "have you tried turning it off and on again"
// expressed in the language of quantum mechanics.
func (r *QuantumHamiltonianReconciler) performQuantumTunneling(
	ctx context.Context,
	h *quantumv1.QuantumHamiltonian,
) error {
	podList := &corev1.PodList{}
	if err := r.List(ctx, podList,
		client.InNamespace(h.Spec.TargetNamespace),
		client.MatchingLabels{"quantum-tiktok-operator": "true"},
	); err != nil {
		return fmt.Errorf("listing pods: %w", err)
	}

	if len(podList.Items) == 0 {
		return nil
	}

	// Select victim by lowest dopamine score annotation.
	// If annotation is absent, select the first pod alphabetically.
	// Both are defensible in a postmortem.
	victim := selectTunnelingVictim(podList.Items)

	log.FromContext(ctx).Info("tunneling pod",
		"pod", victim.Name,
		"node", victim.Spec.NodeName,
	)

	return r.Delete(ctx, &victim)
}

// injectDecoherence fires a Chaos Mesh network failure in a detached goroutine.
// The context is independent to prevent the WiFi-drop command from being
// cancelled by the same network failure it is trying to create.
// This is the distributed systems equivalent of pulling your own power cord.
func (r *QuantumHamiltonianReconciler) injectDecoherence(namespace string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.Chaos.InjectNetworkFailure(ctx, namespace); err != nil {
		// Chaos failed. The system is stable despite our best efforts.
		// Log and move on. Do not page anyone. This is fine.
		log.FromContext(ctx).Info("chaos injection failed — system accidentally stable", "err", err)
	}
}

// setCoherenceState updates the status subresource with the new coherence state
// and appends a standard Kubernetes condition for observability tooling.
func (r *QuantumHamiltonianReconciler) setCoherenceState(
	ctx context.Context,
	h *quantumv1.QuantumHamiltonian,
	state quantumv1.CoherenceState,
) error {
	now := metav1.Now()
	h.Status.CoherenceState = state
	h.Status.LastTransitionTime = &now
	h.Status.ObservedGeneration = h.Generation

	condition := metav1.Condition{
		Type:               conditionTypeCoherent,
		ObservedGeneration: h.Generation,
		LastTransitionTime: now,
	}
	if state == quantumv1.CoherenceStateCoherent {
		condition.Status = metav1.ConditionTrue
		condition.Reason = "OracleConfirmed"
		condition.Message = "Oracle returned ✅"
	} else {
		condition.Status = metav1.ConditionFalse
		condition.Reason = string(state)
		condition.Message = fmt.Sprintf("System entered %s state", state)
	}

	meta.SetStatusCondition(&h.Status.Conditions, condition)

	return r.Status().Update(ctx, h)
}

// SetupWithManager registers the controller and configures watches.
func (r *QuantumHamiltonianReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&quantumv1.QuantumHamiltonian{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}

#!/bin/bash
# ─── Live Environment Test Runner ────────────────────────────
# Tests the Ace Cloud Terraform provider against a real backend.
#
# Usage:
#   ./run-tests.sh plan                # Preview basic resources (safe)
#   ./run-tests.sh apply               # Create basic resources
#   ./run-tests.sh destroy             # Clean up all test resources
#   ./run-tests.sh full                # apply → idempotency → destroy
#
# Update lifecycle tests:
#   ./run-tests.sh update-test         # Full update lifecycle (create → update → idempotency → destroy)
#
# Advanced scenario tests:
#   ./run-tests.sh advanced            # Full advanced test (create → idempotency → destroy)
#
# Run everything:
#   ./run-tests.sh all                 # basic + update + advanced

set -euo pipefail

TERRAFORM="${TERRAFORM:-/Users/akshaymalik/go-sdk/go/bin/terraform}"
DIR="$(cd "$(dirname "$0")" && pwd)"
ACTION="${1:-plan}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

log()  { echo -e "${GREEN}[✓]${NC} $*"; }
warn() { echo -e "${YELLOW}[!]${NC} $*"; }
fail() { echo -e "${RED}[✗]${NC} $*"; exit 1; }
info() { echo -e "${CYAN}[i]${NC} $*"; }

PASS=0
FAIL=0
SKIP=0

check() {
  local name="$1" val="$2"
  if [[ "$val" == "FAILED" || -z "$val" ]]; then
    echo -e "  ${RED}FAIL${NC} $name"
    ((FAIL++))
  elif [[ "$val" == "skipped" ]]; then
    echo -e "  ${YELLOW}SKIP${NC} $name (variable not set)"
    ((SKIP++))
  else
    echo -e "  ${GREEN}PASS${NC} $name → $val"
    ((PASS++))
  fi
}

check_output() {
  local name="$1" output_key="$2"
  local val
  val=$($TERRAFORM output -raw "$output_key" 2>/dev/null || echo "FAILED")
  check "$name" "$val"
}

verify_idempotency() {
  info "Verifying plan idempotency (expect: No changes)..."
  local plan_output
  plan_output=$($TERRAFORM plan -detailed-exitcode "$@" 2>&1) && local exitcode=$? || local exitcode=$?

  if [[ $exitcode -eq 0 ]]; then
    log "Plan idempotency: ${GREEN}PASS${NC} — No changes detected"
    ((PASS++))
  elif [[ $exitcode -eq 2 ]]; then
    echo -e "  ${RED}FAIL${NC} Plan idempotency — Unexpected changes detected:"
    echo "$plan_output" | grep -E "^\s*(~|\\+|-)" | head -20
    ((FAIL++))
  else
    echo -e "  ${RED}FAIL${NC} Plan idempotency — Plan command failed"
    ((FAIL++))
  fi
}

print_results() {
  echo ""
  log "Results: ${GREEN}${PASS} passed${NC}, ${RED}${FAIL} failed${NC}, ${YELLOW}${SKIP} skipped${NC}"
  if [[ $FAIL -gt 0 ]]; then
    fail "Some tests failed!"
  fi
}

# Check prerequisites
[[ -f "$DIR/terraform.tfvars" ]] || fail "Missing terraform.tfvars — copy from terraform.tfvars.example"
command -v "$TERRAFORM" &>/dev/null || fail "Terraform not found at $TERRAFORM"

cd "$DIR"

case "$ACTION" in
  plan)
    log "Running terraform plan..."
    $TERRAFORM plan
    ;;

  apply)
    log "Running terraform apply (basic resources)..."
    $TERRAFORM apply -auto-approve

    echo ""
    log "─── Basic Test Results ────────────────────────────────"

    check_output "VPC"                 "vpc_id"
    check_output "VPC Subnet"          "vpc_subnet_id"
    check_output "Subnet"              "subnet_id"
    check_output "Security Group Web"  "security_group_web_id"
    check_output "Security Group DB"   "security_group_db_id"
    check_output "Key Pair Generated"  "key_pair_generated_id"
    check_output "Key Pair Imported"   "key_pair_imported_id"
    check_output "Volume SSD"          "volume_ssd_id"
    check_output "Volume SSD2"         "volume_ssd2_id"
    check_output "Volume Bootable"     "volume_bootable_id"
    check_output "Router"              "router_id"
    check_output "Router Interface"    "router_interface_id"
    check_output "Snapshot"            "snapshot_id"
    check_output "Volume Backup"       "volume_backup_id"
    check_output "Load Balancer"       "load_balancer_id"
    check_output "LB Listener"         "lb_listener_id"
    check_output "LB Pool"             "lb_pool_id"
    check_output "LB Pool Member"      "lb_pool_member_id"
    check_output "LB Health Monitor"   "lb_health_monitor_id"
    check_output "Instance"            "instance_id"
    check_output "Floating IP"         "floating_ip_id"
    check_output "FIP Association"     "floating_ip_association_id"
    check_output "Auto Scale Template" "auto_scaling_template_id"
    check_output "CaaS Secret"         "caas_secret_id"

    print_results
    ;;

  destroy)
    warn "Destroying all test resources..."
    $TERRAFORM destroy -auto-approve
    log "All test resources destroyed."
    ;;

  full)
    log "Running full basic test cycle: apply → idempotency → destroy"
    echo ""

    # Phase 1: Create
    "$0" apply
    echo ""

    # Phase 2: Idempotency
    verify_idempotency
    echo ""

    # Phase 3: Destroy
    warn "Sleeping 5s before destroy..."
    sleep 5
    "$0" destroy
    echo ""
    print_results
    log "Full basic test cycle complete!"
    ;;

  update-test)
    log "Running update lifecycle test"
    echo ""
    PASS=0; FAIL=0; SKIP=0

    # Phase 1: Create with initial values
    info "Phase 1/4: Creating resources with initial values..."
    $TERRAFORM apply -auto-approve -var="run_update_tests=true" -var="test_phase=create"
    echo ""

    log "─── Create Phase Results ──────────────────────────────"
    check_output "Update VPC"       "update_vpc_name"
    check_output "Update Subnet"    "update_subnet_name"
    check_output "Update SG"        "update_sg_name"
    check_output "Update Router"    "update_router_name"
    check_output "Update Volume"    "update_volume_name"
    check_output "Update Snapshot"  "update_snapshot_name"
    check_output "Update Backup"    "update_backup_name"
    check_output "Update LB"        "update_lb_name"
    check_output "Update HM Delay"  "update_hm_delay"
    check_output "Update Member Wt" "update_member_weight"
    check_output "Update Instance"  "update_instance_name"

    echo ""
    # Phase 2: Idempotency after create
    verify_idempotency -var="run_update_tests=true" -var="test_phase=create"
    echo ""

    # Phase 3: Update to modified values
    info "Phase 3/4: Applying updates (name, description, values)..."
    $TERRAFORM apply -auto-approve -var="run_update_tests=true" -var="test_phase=update"
    echo ""

    log "─── Update Phase Results ──────────────────────────────"
    # Verify values changed
    UPD_VPC=$($TERRAFORM output -raw update_vpc_name 2>/dev/null || echo "FAILED")
    UPD_VOL=$($TERRAFORM output -raw update_volume_name 2>/dev/null || echo "FAILED")
    UPD_VOL_SIZE=$($TERRAFORM output -raw update_volume_size 2>/dev/null || echo "0")
    UPD_HM=$($TERRAFORM output -raw update_hm_delay 2>/dev/null || echo "0")
    UPD_WT=$($TERRAFORM output -raw update_member_weight 2>/dev/null || echo "0")

    # Verify values are the UPDATED values
    if [[ "$UPD_VPC" == *"renamed"* ]]; then
      echo -e "  ${GREEN}PASS${NC} VPC name updated → $UPD_VPC"
      ((PASS++))
    else
      echo -e "  ${RED}FAIL${NC} VPC name not updated (expected *renamed*, got $UPD_VPC)"
      ((FAIL++))
    fi

    if [[ "$UPD_VOL" == *"renamed"* ]]; then
      echo -e "  ${GREEN}PASS${NC} Volume name updated → $UPD_VOL"
      ((PASS++))
    else
      echo -e "  ${RED}FAIL${NC} Volume name not updated (expected *renamed*, got $UPD_VOL)"
      ((FAIL++))
    fi

    if [[ "$UPD_VOL_SIZE" == "15" ]]; then
      echo -e "  ${GREEN}PASS${NC} Volume size updated → ${UPD_VOL_SIZE}GB"
      ((PASS++))
    else
      echo -e "  ${RED}FAIL${NC} Volume size not updated (expected 15, got $UPD_VOL_SIZE)"
      ((FAIL++))
    fi

    if [[ "$UPD_HM" == "15" ]]; then
      echo -e "  ${GREEN}PASS${NC} HM delay updated → ${UPD_HM}s"
      ((PASS++))
    else
      echo -e "  ${RED}FAIL${NC} HM delay not updated (expected 15, got $UPD_HM)"
      ((FAIL++))
    fi

    if [[ "$UPD_WT" == "5" ]]; then
      echo -e "  ${GREEN}PASS${NC} Pool member weight updated → $UPD_WT"
      ((PASS++))
    else
      echo -e "  ${RED}FAIL${NC} Pool member weight not updated (expected 5, got $UPD_WT)"
      ((FAIL++))
    fi

    echo ""
    # Phase 4: Idempotency after update
    verify_idempotency -var="run_update_tests=true" -var="test_phase=update"
    echo ""

    # Phase 5: Destroy
    warn "Sleeping 5s before destroy..."
    sleep 5
    $TERRAFORM destroy -auto-approve -var="run_update_tests=true"
    log "Update test resources destroyed."
    echo ""
    print_results
    log "Update lifecycle test complete!"
    ;;

  advanced)
    log "Running advanced scenario tests"
    echo ""
    PASS=0; FAIL=0; SKIP=0

    # Phase 1: Create
    info "Phase 1/3: Creating advanced resources..."
    $TERRAFORM apply -auto-approve -var="run_advanced_tests=true"
    echo ""

    log "─── Advanced Test Results ─────────────────────────────"
    check_output "Multi-Subnet VPC"      "adv_vpc_id"
    check_output "App Subnet"            "adv_subnet_app_id"
    check_output "DB Subnet"             "adv_subnet_db_id"
    check_output "Router + Gateway"      "adv_router_id"
    check_output "Boot Volume"           "adv_boot_volume_id"
    check_output "Boot-from-Vol Instance" "adv_instance_id"
    check_output "Floating IP"           "adv_floating_ip"
    check_output "Multi-Member LB"       "adv_lb_id"
    check_output "TCP LB"                "adv_tcp_lb_id"
    check_output "Minimal VPC"           "adv_minimal_vpc_id"
    check_output "Minimal SG"            "adv_minimal_sg_id"
    check_output "Snapshot"              "adv_snapshot_id"
    check_output "Backup"                "adv_backup_id"

    echo ""
    # Phase 2: Idempotency
    verify_idempotency -var="run_advanced_tests=true"
    echo ""

    # Phase 3: Destroy
    warn "Sleeping 5s before destroy..."
    sleep 5
    $TERRAFORM destroy -auto-approve -var="run_advanced_tests=true"
    log "Advanced test resources destroyed."
    echo ""
    print_results
    log "Advanced scenario tests complete!"
    ;;

  all)
    log "Running ALL test suites"
    echo ""
    echo "═══════════════════════════════════════════════════════"
    echo " Suite 1: Basic Resources"
    echo "═══════════════════════════════════════════════════════"
    "$0" full
    echo ""
    echo "═══════════════════════════════════════════════════════"
    echo " Suite 2: Update Lifecycle"
    echo "═══════════════════════════════════════════════════════"
    "$0" update-test
    echo ""
    echo "═══════════════════════════════════════════════════════"
    echo " Suite 3: Advanced Scenarios"
    echo "═══════════════════════════════════════════════════════"
    "$0" advanced
    echo ""
    log "ALL test suites complete!"
    ;;

  *)
    echo "Usage: $0 {plan|apply|destroy|full|update-test|advanced|all}"
    echo ""
    echo "  plan         Preview basic resources (safe)"
    echo "  apply        Create basic resources"
    echo "  destroy      Clean up all resources"
    echo "  full         Basic lifecycle: apply → idempotency → destroy"
    echo "  update-test  Update lifecycle: create → update → idempotency → destroy"
    echo "  advanced     Advanced scenarios: multi-subnet, boot-from-vol, TCP LB, etc."
    echo "  all          Run all 3 test suites sequentially"
    exit 1
    ;;
esac

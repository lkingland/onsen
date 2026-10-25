#!/bin/bash

set -euo pipefail

source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/base-test.sh"

rule="$ROOT/etc/udev/rules.d/70-onsen-bluetooth-link-policy.rules"
command_file="$ROOT/bin/omarchy-bluetooth-link-policy"
migration="$ROOT/migrations/1792891069.sh"

[[ -f $rule ]] || fail "the link-policy udev rule is shipped"
grep -q 'RUN+="/usr/bin/omarchy-bluetooth-link-policy \$attr{uniq}"' "$rule" || \
  fail "the link-policy rule runs the shipped command with the device address"
grep -q 'KERNEL=="input\[0-9\]\*"' "$rule" || \
  fail "the link-policy rule matches the input device once, not its event node too"
pass "the link-policy rule runs the shipped command once per trackpad link"

[[ -x $command_file ]] || fail "omarchy-bluetooth-link-policy is executable"
python3 -m py_compile "$command_file" 2>/dev/null || fail "omarchy-bluetooth-link-policy parses"
grep -q '^# omarchy:hidden=true$' "$command_file" || fail "omarchy-bluetooth-link-policy is hidden from the command listing"
pass "omarchy-bluetooth-link-policy is a hidden, valid command"

grep -q '^options btusb enable_autosuspend=0$' "$ROOT/etc/modprobe.d/bluetooth.conf" || \
  fail "the packaged modprobe file turns off btusb autosuspend"
! grep -q '^options bluetooth disable_ertm' "$ROOT/etc/modprobe.d/bluetooth.conf" || fail "the retired ERTM option is gone"
[[ ! -e $ROOT/etc/udev/rules.d/70-onsen-bluetooth-connectable.rules ]] || fail "the retired connectable rule is gone"
pass "the dongle fix is the one modprobe option and the retired fixes are gone"

[[ -f $migration ]] || fail "the migration retiring the hand-made dongle files exists"
grep -q 'btusb-no-autosuspend.conf' "$migration" && grep -q '50-bluetooth-dongle-no-autosuspend.rules' "$migration" || \
  fail "the migration names both hand-made files"
grep -q 'omarchy-bluetooth-link-policy' "$migration" || fail "the migration sets the policy on a trackpad already connected"
pass "the migration retires the hand-made files and applies the policy"

#!/usr/bin/env bash
set -euo pipefail
set -x

# const
SED="/usr/bin/sed"
NONE=0

# tunable
IRQBALANCE_BANNED_CPUS="${IRQBALANCE_BANNED_CPUS:-${NONE}}"

# tunable - overridable for testing purposes
IRQBALANCE_CONF="${1:-/etc/sysconfig/irqbalance}"
CRIO_ORIG_BANNED_CPUS="${2:-/etc/sysconfig/orig_irq_banned_cpus}"
DEFAULT_SMP_AFFINITY_CONF="${3:-/proc/irq/default_smp_affinity}"

[ ! -f "${IRQBALANCE_CONF}" ] && exit 0

if [ "$IRQBALANCE_BANNED_CPUS" -ne "$NONE" ]; then
	default_smp_affinity=$(cat "$DEFAULT_SMP_AFFINITY_CONF")
	printf '%x\n' $(("0x$default_smp_affinity" & ~ "0x$IRQBALANCE_BANNED_CPUS")) > "$DEFAULT_SMP_AFFINITY_CONF"
fi

${SED} -i '/^\s*IRQBALANCE_BANNED_CPUS\b/d' "${IRQBALANCE_CONF}" || exit 0
# CPU numbers which have their corresponding bits set to one in this mask
# will not have any irq's assigned to them on rebalance.
# so zero means all cpus are participating in load balancing.
echo "IRQBALANCE_BANNED_CPUS=${IRQBALANCE_BANNED_CPUS}" >> "${IRQBALANCE_CONF}"

# we now own this configuration. But CRI-O has code to restore the configuration,
# and until it gains the option to disable this restore flow, we need to make
# the configuration consistent such as the CRI-O restore will do nothing.
if [ -n "${CRIO_ORIG_BANNED_CPUS}" ] && [ -f "${CRIO_ORIG_BANNED_CPUS}" ]; then
	echo "${IRQBALANCE_BANNED_CPUS}" > "${CRIO_ORIG_BANNED_CPUS}"
fi

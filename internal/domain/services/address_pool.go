package services

import (
	"fmt"
	"net/netip"

	"github.com/thilob/wg-peerforge/internal/domain/models"
)

func NextAvailableIPv4(pool models.AddressPool) (string, error) {
	prefix, err := netip.ParsePrefix(pool.CIDR)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR %q: %w", pool.CIDR, err)
	}

	if !prefix.Addr().Is4() {
		return "", fmt.Errorf("only IPv4 pools are supported in v1")
	}

	reserved := map[string]struct{}{}
	for _, address := range pool.Allocated {
		addrPrefix, err := netip.ParsePrefix(address)
		if err != nil {
			return "", fmt.Errorf("invalid allocated address %q: %w", address, err)
		}
		reserved[addrPrefix.Addr().String()] = struct{}{}
	}

	base := prefix.Masked().Addr()
	for candidate := base.Next(); prefix.Contains(candidate); candidate = candidate.Next() {
		last, ok := lastAddress(prefix)
		if !ok {
			break
		}
		if candidate == last {
			break
		}
		if _, exists := reserved[candidate.String()]; exists {
			continue
		}
		return netip.PrefixFrom(candidate, 32).String(), nil
	}

	return "", fmt.Errorf("no free IPv4 addresses left in %s", pool.CIDR)
}

func lastAddress(prefix netip.Prefix) (netip.Addr, bool) {
	addr := prefix.Masked().Addr()
	if !addr.Is4() {
		return netip.Addr{}, false
	}

	bytes := addr.As4()
	hostBits := 32 - prefix.Bits()
	for i := range 4 {
		bitsForOctet := hostBits - ((3 - i) * 8)
		switch {
		case bitsForOctet >= 8:
			bytes[i] = 255
		case bitsForOctet > 0:
			bytes[i] |= byte((1 << bitsForOctet) - 1)
		}
	}

	return netip.AddrFrom4(bytes), true
}

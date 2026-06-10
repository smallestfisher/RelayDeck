package router

import (
	"errors"
	"math"
	"slices"

	"github.com/smallestfisher/relaydeck/backend/internal/domain"
)

var ErrNoCandidate = errors.New("no upstream candidate available")

func SelectCandidate(
	req domain.GatewayRequest,
	_ domain.GatewayPrincipal,
	sites []domain.UpstreamSite,
	mappings []domain.SiteModel,
	policy domain.RoutingPolicy,
) (domain.RouteCandidate, error) {
	var best domain.RouteCandidate
	found := false

	for _, site := range sites {
		if !site.Enabled || site.Circuit == domain.CircuitOpen || site.HealthScore < policy.MinimumHealthScore {
			continue
		}
		mapping, ok := mappingForSite(req, site, mappings)
		if !ok {
			continue
		}
		candidate := domain.RouteCandidate{Site: site, Mapping: mapping, Score: score(site)}
		if !found || candidate.Score > best.Score {
			best = candidate
			found = true
		}
	}

	if !found {
		return domain.RouteCandidate{}, ErrNoCandidate
	}
	return best, nil
}

func mappingForSite(req domain.GatewayRequest, site domain.UpstreamSite, mappings []domain.SiteModel) (domain.SiteModel, bool) {
	for _, mapping := range mappings {
		if mapping.SiteID != site.ID || mapping.Model != req.Model {
			continue
		}
		if !slices.Contains(mapping.EndpointTypes, req.Endpoint) {
			continue
		}
		if !hasCapabilities(site.Capabilities, req.RequiredCapabilities) || !hasCapabilities(mapping.Capabilities, req.RequiredCapabilities) {
			continue
		}
		return mapping, true
	}
	return domain.SiteModel{}, false
}

func hasCapabilities(available []domain.Capability, required []domain.Capability) bool {
	for _, capability := range required {
		if !slices.Contains(available, capability) {
			return false
		}
	}
	return true
}

func score(site domain.UpstreamSite) float64 {
	latencyScore := math.Max(0, 100-float64(site.LatencyMS)/10)
	weight := site.Weight
	if weight == 0 {
		weight = 50
	}
	return site.HealthScore*0.40 + site.SuccessRate*0.25 + latencyScore*0.20 + weight*0.10 + 100*0.05
}

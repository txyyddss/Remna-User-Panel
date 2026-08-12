package admin

import "context"

func (s *SettingsService) paymentProfileReadiness(ctx context.Context, provider string) ([]string, bool) {
	if s.profiles == nil {
		return nil, false
	}
	profiles, err := s.profiles.ListPaymentProfiles(ctx)
	if err != nil || len(profiles) == 0 {
		return nil, false
	}
	for _, profile := range profiles {
		if profile.Provider != provider {
			continue
		}
		if !profile.Enabled {
			return nil, true
		}
		issues := make([]string, 0, 2)
		if !profile.Configured || len(profile.EnabledChannels) == 0 {
			issues = append(issues, "missing:payment_profile."+provider)
		}
		if provider == "ezpay" && profile.MerchantID == "" {
			issues = append(issues, "missing:payment_profile."+provider+".merchant_id")
		}
		return issues, true
	}
	return nil, false
}

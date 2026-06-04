package handlers

import (
	"math"
	"net/http"
	"time"

	"syntheticvision/internal/middleware"
	"syntheticvision/internal/models"
)

type AnalyticsSummary struct {
	TotalGenerations      int64   `json:"total_generations"`
	CompletedGenerations  int64   `json:"completed_generations"`
	FailedGenerations     int64   `json:"failed_generations"`
	PendingGenerations    int64   `json:"pending_generations"`
	ProcessingGenerations int64   `json:"processing_generations"`
	SuccessRate           float64 `json:"success_rate"`
	CreditsSpent          int     `json:"credits_spent"`
	CreditsRefunded       int     `json:"credits_refunded"`
	CreditsBalance        int     `json:"credit_balance"`
}

type AnalyticsDistributionItem struct {
	Label      string  `json:"label"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

type AnalyticsRecentGeneration struct {
	ID          string     `json:"id"`
	Prompt      string     `json:"prompt"`
	Status      string     `json:"status"`
	Resolution  string     `json:"resolution"`
	AspectRatio string     `json:"aspect_ratio"`
	Cost        int        `json:"cost"`
	Error       string     `json:"error"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type AnalyticsCreditBreakdown struct {
	GenerationDebits int `json:"generation_debits"`
	Refunds          int `json:"refunds"`
	AdminTopups      int `json:"admin_topups"`
	SignupBonus      int `json:"signup_bonus"`
}

type AnalyticsResponse struct {
	User                   models.User                 `json:"user"`
	Summary                AnalyticsSummary            `json:"summary"`
	StatusDistribution     []AnalyticsDistributionItem `json:"status_distribution"`
	ResolutionDistribution []AnalyticsDistributionItem `json:"resolution_distribution"`
	AspectDistribution     []AnalyticsDistributionItem `json:"aspect_ratio_distribution"`
	CreditBreakdown        AnalyticsCreditBreakdown    `json:"credit_breakdown"`
	RecentGenerations      []AnalyticsRecentGeneration `json:"recent_generations"`
}

// Analytics returns the authenticated user's personal metrics derived from generation
// and credit rows.
//
//		GET /api/me/analytics -> {
//		 user: User,
//		 summary: {...},
//		 status_distribution: [...],
//		 resolution_distribution: [...],
//		 aspect_ratio_distribution:[...],
//		 credit_breakdown: {...},
//		 recent_generations: [...]
//	}
func (h *Handler) Analytics(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r)

	var user models.User
	if err := h.DB.Where("id = ?", uid).First(&user).Error; err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	var summary AnalyticsSummary
	summary.CreditsBalance = user.Credits

	var total int64
	if err := h.DB.Model(&models.Generation{}).Where("user_id = ?", uid).Count(&total).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to aggregate analytics")
		return
	}
	summary.TotalGenerations = total

	var completed int64
	if err := h.DB.Model(&models.Generation{}).
		Where("user_id = ? AND status = ?", uid, "completed").Count(&completed).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to aggregate analytics")
		return
	}
	summary.CompletedGenerations = completed

	var failed int64
	if err := h.DB.Model(&models.Generation{}).
		Where("user_id = ? AND status = ?", uid, "failed").Count(&failed).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to aggregate analytics")
		return
	}
	summary.FailedGenerations = failed

	var pending int64
	if err := h.DB.Model(&models.Generation{}).
		Where("user_id = ? AND status = ?", uid, "pending").Count(&pending).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to aggregate analytics")
		return
	}
	summary.PendingGenerations = pending

	var processing int64
	if err := h.DB.Model(&models.Generation{}).
		Where("user_id = ? AND status = ?", uid, "processing").Count(&processing).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to aggregate analytics")
		return
	}
	summary.ProcessingGenerations = processing

	if total == 0 {
		summary.SuccessRate = 0
	} else {
		summary.SuccessRate = round2(float64(completed) * 100 / float64(total))
	}

	var creditRows []struct {
		Reason string `gorm:"column:reason"`
		Total  int64  `gorm:"column:total"`
	}
	if err := h.DB.Model(&models.CreditTransaction{}).
		Select("reason, COALESCE(SUM(amount), 0) AS total").
		Where("user_id = ?", uid).
		Group("reason").
		Find(&creditRows).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to aggregate analytics")
		return
	}

	var breakdown AnalyticsCreditBreakdown
	for _, row := range creditRows {
		switch row.Reason {
		case "generation":
			if row.Total < 0 {
				breakdown.GenerationDebits += int(-row.Total)
			}
		case "refund":
			if row.Total > 0 {
				breakdown.Refunds += int(row.Total)
			}
		case "admin_injection":
			breakdown.AdminTopups += int(row.Total)
		case "signup_bonus":
			breakdown.SignupBonus += int(row.Total)
		}
	}
	summary.CreditsSpent = breakdown.GenerationDebits
	summary.CreditsRefunded = breakdown.Refunds

	statusDistribution, err := buildStatusDistribution(h, uid, summary.TotalGenerations)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to aggregate analytics")
		return
	}
	resolutionDistribution, err := buildResolutionDistribution(h, uid, summary.TotalGenerations)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to aggregate analytics")
		return
	}
	aspectDistribution, err := buildAspectDistribution(h, uid, summary.TotalGenerations)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to aggregate analytics")
		return
	}

	var recentGenerations []models.Generation
	if err := h.DB.Where("user_id = ?", uid).Order("created_at DESC").Limit(10).Find(&recentGenerations).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to aggregate analytics")
		return
	}

	recent := make([]AnalyticsRecentGeneration, 0, len(recentGenerations))
	for _, gen := range recentGenerations {
		recent = append(recent, AnalyticsRecentGeneration{
			ID:          gen.ID,
			Prompt:      gen.Prompt,
			Status:      gen.Status,
			Resolution:  gen.Resolution,
			AspectRatio: gen.AspectRatio,
			Cost:        gen.Cost,
			Error:       gen.Error,
			CreatedAt:   gen.CreatedAt,
			CompletedAt: gen.CompletedAt,
		})
	}

	writeJSON(w, http.StatusOK, AnalyticsResponse{
		User:                   user,
		Summary:                summary,
		StatusDistribution:     statusDistribution,
		ResolutionDistribution: resolutionDistribution,
		AspectDistribution:     aspectDistribution,
		CreditBreakdown:        breakdown,
		RecentGenerations:      recent,
	})
}

func buildStatusDistribution(h *Handler, uid string, total int64) ([]AnalyticsDistributionItem, error) {
	type row struct {
		Status string `gorm:"column:status"`
		Count  int64  `gorm:"column:count"`
	}
	rows := []row{}
	if err := h.DB.Model(&models.Generation{}).
		Select("status, COUNT(*) AS count").
		Where("user_id = ?", uid).
		Group("status").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	counts := map[string]int64{
		"pending":    0,
		"processing": 0,
		"completed":  0,
		"failed":     0,
	}
	for _, r := range rows {
		counts[r.Status] += r.Count
	}

	items := make([]AnalyticsDistributionItem, 0, 4)
	for _, status := range []string{"completed", "failed", "pending", "processing"} {
		items = append(items, AnalyticsDistributionItem{
			Label:      status,
			Count:      counts[status],
			Percentage: percent(counts[status], total),
		})
	}
	return items, nil
}

func buildResolutionDistribution(h *Handler, uid string, total int64) ([]AnalyticsDistributionItem, error) {
	type row struct {
		Resolution string `gorm:"column:resolution"`
		Count      int64  `gorm:"column:count"`
	}
	rows := []row{}
	if err := h.DB.Model(&models.Generation{}).
		Select("resolution, COUNT(*) AS count").
		Where("user_id = ?", uid).
		Group("resolution").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	counts := map[string]int64{}
	for _, r := range rows {
		counts[r.Resolution] += r.Count
	}
	// Keep display ordering stable with product-supported values.
	ordered := []string{"1K", "2K", "4K"}
	items := make([]AnalyticsDistributionItem, 0, len(ordered))
	for _, res := range ordered {
		items = append(items, AnalyticsDistributionItem{
			Label:      res,
			Count:      counts[res],
			Percentage: percent(counts[res], total),
		})
	}
	return items, nil
}

func buildAspectDistribution(h *Handler, uid string, total int64) ([]AnalyticsDistributionItem, error) {
	type row struct {
		AspectRatio string `gorm:"column:aspect_ratio"`
		Count       int64  `gorm:"column:count"`
	}
	rows := []row{}
	if err := h.DB.Model(&models.Generation{}).
		Select("aspect_ratio, COUNT(*) AS count").
		Where("user_id = ?", uid).
		Group("aspect_ratio").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	counts := map[string]int64{}
	for _, r := range rows {
		counts[r.AspectRatio] += r.Count
	}
	ordered := []string{"1:1", "4:3", "16:9", "9:16"}
	items := make([]AnalyticsDistributionItem, 0, len(ordered))
	for _, ratio := range ordered {
		items = append(items, AnalyticsDistributionItem{
			Label:      ratio,
			Count:      counts[ratio],
			Percentage: percent(counts[ratio], total),
		})
	}
	return items, nil
}

func percent(part, total int64) float64 {
	if total == 0 {
		return 0
	}
	return round2(float64(part) * 100 / float64(total))
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

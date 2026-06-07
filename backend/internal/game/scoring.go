package game

import "math"

const (
	StreakBonusThreshold = 3    // Nechta ketma-ket to'g'ri javobdan bonus boshlanadi
	StreakBonusPercent   = 10   // Har streak uchun foiz qo'shimcha
	MaxStreakBonus       = 50   // Maksimal streak bonus foizi
	EarlyBirdBonusMax   = 50   // Birinchi javob beruvchi bonusi (max)
)

// CalcPoints hisoblaydi: asosiy ball + tezlik bonusi + streak bonusi
// basePoints: savolga belgilangan maksimal ball
// timeLimitSec: savolga berilgan vaqt (soniya)
// responseTimeSec: o'quvchi javob bergan vaqt (soniya)
// streak: ketma-ket to'g'ri javoblar soni
// answerIndex: nechtanchi o'quvchi javob berdi (early-bird uchun)
func CalcPoints(basePoints, timeLimitSec, responseTimeSec, streak, answerIndex int) (total, speedBonus, streakBonus int) {
	if responseTimeSec > timeLimitSec {
		responseTimeSec = timeLimitSec
	}

	// Tezlik bonusi: vaqt qanchalik kam bo'lsa, shuncha ko'p ball
	// Formula: basePoints * (1 - responseTime/timeLimit * 0.5)
	// Eng tez javob = 100%, eng kech javob = 50%
	speedFactor := 1.0 - (float64(responseTimeSec)/float64(timeLimitSec))*0.5
	speedBonus = int(math.Round(float64(basePoints) * speedFactor))

	// Streak bonusi
	if streak >= StreakBonusThreshold {
		bonusPercent := min(streak*StreakBonusPercent/StreakBonusThreshold, MaxStreakBonus)
		streakBonus = int(math.Round(float64(speedBonus) * float64(bonusPercent) / 100))
	}

	total = speedBonus + streakBonus
	return
}

// CalcAccuracyPoints — Accuracy Mode: tezlikdan qat'iy nazar, to'g'ri javob = to'liq ball
func CalcAccuracyPoints(basePoints int) int {
	return basePoints
}

// CalcConfidencePoints — Confidence Mode
// level: 1=unsure(x0.5), 2=maybe(x1.0), 3=sure(x2.0)
// isCorrect: to'g'ri yoki noto'g'ri
func CalcConfidencePoints(basePoints, level int, isCorrect bool) int {
	multipliers := map[int]float64{1: 0.5, 2: 1.0, 3: 2.0}
	m, ok := multipliers[level]
	if !ok {
		m = 1.0
	}
	if isCorrect {
		return int(math.Round(float64(basePoints) * m))
	}
	// Noto'g'ri javob uchun jarima (sure deb o'ylagan bo'lsa)
	if level == 3 {
		return -int(math.Round(float64(basePoints) * 0.5))
	}
	return 0
}

// UpdateLeaderboard sorts participants by score and assigns ranks
func UpdateLeaderboard(participants []*ParticipantState) {
	// Simple insertion sort (kichik sonlar uchun yetarli)
	for i := 1; i < len(participants); i++ {
		key := participants[i]
		j := i - 1
		for j >= 0 && participants[j].Score < key.Score {
			participants[j+1] = participants[j]
			j--
		}
		participants[j+1] = key
	}

	for i, p := range participants {
		prevRank := p.Rank
		p.Rank = i + 1
		p.RankDelta = prevRank - p.Rank // positive = moved up
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

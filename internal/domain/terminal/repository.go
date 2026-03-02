package terminal

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

// Repository - интерфейс (контракт), по которому мы работаем с данными
type Repository interface {
	EnrichATM(atm *ATM)
	EnrichCompetitor(atm *ATM)
	GenerateRandomCompetitors(count int) []ATM
}

// MockRepository - имитация базы данных
type MockRepository struct{}

func NewMockRepository() *MockRepository {
	return &MockRepository{}
}

// --- 1. ЛОГИКА ДЛЯ FORTE (Детальная) ---
func (r *MockRepository) EnrichATM(atm *ATM) {
	atm.IsForte = true
	atm.Bank = "Forte Bank"
	atm.District = "Город"

	atm.AvgCashBalanceKZT = float64(5000000 + rand.Intn(20000000))
	atm.WithdrawalFreqPerDay = 50 + rand.Intn(400)
	atm.DowntimePct = rand.Float64() * 0.15

	var total float64
	atm.Cassettes = r.genCassettes()
	// Считаем сумму
	for _, c := range atm.Cassettes {
		total += c.Amount
	}
	atm.TotalCashKZT = total

	// 👇 ТЕПЕРЬ ЭТО ВОЗВРАЩАЕТ СТРОКУ (JSON)
	atm.Complaints = r.genComplaints()

	r.calcEfficiency(atm)
}

// --- 2. ЛОГИКА ДЛЯ КОНКУРЕНТОВ (Оценочная) ---
func (r *MockRepository) EnrichCompetitor(atm *ATM) {
	atm.IsForte = false
	atm.EstWithdrawalKZT = float64(2000000 + rand.Intn(13000000))
	atm.EstDepositKZT = float64(500000 + rand.Intn(7500000))
}

// --- 3. FALLBACK ГЕНЕРАТОР ---
func (r *MockRepository) GenerateRandomCompetitors(count int) []ATM {
	var atms []ATM
	banks := []string{"Kaspi", "Halyk", "Jusan", "BCC", "Eurasian"}
	minLat, maxLat := 51.08, 51.20
	minLng, maxLng := 71.38, 71.52

	for i := 0; i < count; i++ {
		bank := banks[rand.Intn(len(banks))]
		atms = append(atms, ATM{
			ID:               9000 + i,
			Name:             fmt.Sprintf("%s ATM #%d", bank, i),
			Lat:              minLat + rand.Float64()*(maxLat-minLat),
			Lng:              minLng + rand.Float64()*(maxLng-minLng),
			IsForte:          false,
			Bank:             bank,
			EstWithdrawalKZT: float64(2000000 + rand.Intn(10000000)),
			EstDepositKZT:    float64(1000000 + rand.Intn(5000000)),
		})
	}
	return atms
}

// --- ВСПОМОГАТЕЛЬНЫЕ МЕТОДЫ ---

func (r *MockRepository) genCassettes() []Cassette {
	capOut := 20000000.0
	amtOut := float64(rand.Intn(int(capOut)))
	stOut := "OK"
	if amtOut < 2000000 {
		stOut = "Low"
	}
	if amtOut == 0 {
		stOut = "Empty"
	}

	capIn := 10000000.0
	amtIn := float64(rand.Intn(int(capIn)))
	stIn := "OK"
	if amtIn > 9000000 {
		stIn = "Full"
	}

	return []Cassette{
		{Type: "Cash-Out", Currency: "KZT", Amount: amtOut, Capacity: capOut, Status: stOut},
		{Type: "Cash-In", Currency: "KZT", Amount: amtIn, Capacity: capIn, Status: stIn},
	}
}

// 👇 ИСПРАВЛЕННЫЙ МЕТОД: Возвращает string (JSON), а не []Complaint
func (r *MockRepository) genComplaints() string {
	// У 70% терминалов жалоб нет -> возвращаем пустой JSON массив
	if rand.Float32() > 0.3 {
		return "[]"
	}

	templates := []struct{ cat, txt string }{
		{"Техническая", "Зажевал карту"},
		{"Техническая", "Не выдал чек"},
		{"Чистота", "Грязная клавиатура"},
		{"Обслуживание", "Долго обрабатывает запрос"},
	}

	count := 1 + rand.Intn(2)
	var list []Complaint

	for i := 0; i < count; i++ {
		t := templates[rand.Intn(len(templates))]
		list = append(list, Complaint{
			ID:       rand.Intn(99999),
			Category: t.cat,
			Text:     t.txt,
			Status:   "Open",
			Date:     time.Now().AddDate(0, 0, -rand.Intn(10)).Format("2006-01-02"),
		})
	}

	// Сериализуем в JSON строку
	bytes, _ := json.Marshal(list)
	return string(bytes)
}

func (r *MockRepository) calcEfficiency(atm *ATM) {
	if len(atm.Cassettes) < 2 {
		atm.EfficiencyStatus = "Normal"
		return
	}
	cashOutStatus := atm.Cassettes[0].Status
	cashInStatus := atm.Cassettes[1].Status

	if cashOutStatus == "Empty" || cashInStatus == "Full" || atm.DowntimePct > 0.10 {
		atm.EfficiencyStatus = "Ineffective"
	} else if atm.WithdrawalFreqPerDay > 300 && atm.DowntimePct < 0.03 {
		atm.EfficiencyStatus = "Effective"
	} else {
		atm.EfficiencyStatus = "Normal"
	}
}

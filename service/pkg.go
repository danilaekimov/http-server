package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
)

// Service структура для хранения статистики голосования
type Service struct {
	mu    sync.RWMutex
	stats map[uint]uint // Карта, где ключ - ID кандидата, значение - количество голосов
}

// AddVote добавляет голос за указанного кандидата
func (s *Service) AddVote(candidateID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.stats[candidateID]; !ok {
		s.stats[candidateID] = 0
	}
	s.stats[candidateID]++
}

// GetStats возвращает текущую статистику голосования
func (s *Service) GetStats() map[uint]uint {
	return s.stats
}

// GetStatsForCandidate возвращает статистику для конкретного кандидата
func (s *Service) GetStatsForCandidate(candidateID uint) (uint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if votes, ok := s.stats[candidateID]; ok {
		return votes, nil
	}
	return 0, fmt.Errorf("no data for candidate with id %d", candidateID)
}

// VoteHandler обрабатывает запрос на голосование
func (s *Service) VoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	req := struct {
		CandidateID uint   `json:"candidate_id"`
		Passport    string `json:"passport"`
	}{}

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(raw, &req); err != nil || req.CandidateID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.AddVote(req.CandidateID)

	w.WriteHeader(http.StatusOK)
}

// StatsHandler возвращает текущую статистику голосования
func (s *Service) StatsHandler(w http.ResponseWriter, r *http.Request) {
	candidateIDStr := r.URL.Query().Get("candidate_id") // Получаем параметр candidate_id из URL
	var votes uint
	var err error

	if candidateIDStr == "" { // Если нет параметра candidate_id, возвращаем общую статистику
		stats := s.GetStats()
		jsonData, err := json.Marshal(stats)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonData)
		if err != nil {
			fmt.Printf("Failed to write response body: %v\n", err)
		}
		return
	}

	// Преобразуем строку в число
	candidateID, err := strconv.ParseUint(candidateIDStr, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Получаем статистику для конкретного кандидата
	votes, err = s.GetStatsForCandidate(uint(candidateID))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Формируем JSON-ответ
	response := struct {
		CandidateID uint `json:"candidate_id"`
		Votes       uint `json:"votes"`
	}{
		CandidateID: uint(candidateID),
		Votes:       votes,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonData)
	if err != nil {
		fmt.Printf("Failed to write response body: %v\n", err)
	}
}

// NewService создаёт новый экземпляр сервиса
func NewService() *Service {
	return &Service{
		stats: make(map[uint]uint),
	}
}

package download

import (
	"errors"
	"fmt"
	"sync"
)

type PieceStatus int

const (
	PiecePending PieceStatus = iota
	PieceInProgress
	PieceVerified
	PieceFailed
)

type PieceTask struct {
	Index  int
	Length int
	Hash   [20]byte
}

type PieceResult struct {
	Index int
	Data  []byte
}

type Scheduler struct {
	mu     sync.Mutex
	order  []int
	pieces map[int]*pieceState
}

type pieceState struct {
	task   PieceTask
	status PieceStatus
}

func NewScheduler(tasks []PieceTask) (*Scheduler, error) {
	pieces := make(map[int]*pieceState, len(tasks))
	order := make([]int, 0, len(tasks))
	for _, task := range tasks {
		if task.Index < 0 {
			return nil, fmt.Errorf("piece index cannot be negative: %d", task.Index)
		}
		if task.Length <= 0 {
			return nil, fmt.Errorf("piece %d length must be positive", task.Index)
		}
		if _, exists := pieces[task.Index]; exists {
			return nil, fmt.Errorf("duplicate piece index %d", task.Index)
		}
		pieces[task.Index] = &pieceState{task: task, status: PiecePending}
		order = append(order, task.Index)
	}

	return &Scheduler{order: order, pieces: pieces}, nil
}

// NextPiece returns pending work that the peer can serve.
func (s *Scheduler) NextPiece(hasPiece func(index int) bool) (PieceTask, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if hasPiece == nil {
		hasPiece = func(int) bool { return true }
	}

	for _, index := range s.order {
		piece := s.pieces[index]
		if piece.status != PiecePending && piece.status != PieceFailed {
			continue
		}
		if !hasPiece(index) {
			continue
		}
		piece.status = PieceInProgress
		return piece.task, true
	}
	return PieceTask{}, false
}

func (s *Scheduler) Complete(index int) error {
	return s.setStatus(index, PieceVerified)
}

func (s *Scheduler) Retry(index int) error {
	return s.setStatus(index, PiecePending)
}

func (s *Scheduler) Fail(index int) error {
	return s.setStatus(index, PieceFailed)
}

func (s *Scheduler) Status(index int) (PieceStatus, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	piece, ok := s.pieces[index]
	if !ok {
		return PiecePending, false
	}
	return piece.status, true
}

func (s *Scheduler) Progress() (verified int, total int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, piece := range s.pieces {
		if piece.status == PieceVerified {
			verified++
		}
	}
	return verified, len(s.pieces)
}

func (s *Scheduler) Done() bool {
	verified, total := s.Progress()
	return total > 0 && verified == total
}

func (s *Scheduler) setStatus(index int, status PieceStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	piece, ok := s.pieces[index]
	if !ok {
		return errors.New("unknown piece index")
	}
	piece.status = status
	return nil
}

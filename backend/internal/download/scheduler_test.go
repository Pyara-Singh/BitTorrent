package download

import "testing"

func TestSchedulerAssignsOnlyAvailablePieces(t *testing.T) {
	scheduler, err := NewScheduler([]PieceTask{
		{Index: 0, Length: 10},
		{Index: 1, Length: 10},
	})
	if err != nil {
		t.Fatalf("NewScheduler() failed: %v", err)
	}

	task, ok := scheduler.NextPiece(func(index int) bool { return index == 1 })
	if !ok {
		t.Fatal("expected task")
	}
	if task.Index != 1 {
		t.Fatalf("task index = %d, want 1", task.Index)
	}

	status, ok := scheduler.Status(1)
	if !ok || status != PieceInProgress {
		t.Fatalf("status = %v, ok = %v", status, ok)
	}
}

func TestSchedulerRetryAndProgress(t *testing.T) {
	scheduler, err := NewScheduler([]PieceTask{{Index: 0, Length: 10}})
	if err != nil {
		t.Fatalf("NewScheduler() failed: %v", err)
	}

	task, ok := scheduler.NextPiece(nil)
	if !ok || task.Index != 0 {
		t.Fatalf("unexpected first task: %+v ok=%v", task, ok)
	}
	if _, ok := scheduler.NextPiece(nil); ok {
		t.Fatal("in-progress task should not be reassigned")
	}
	if err := scheduler.Fail(0); err != nil {
		t.Fatalf("Fail() failed: %v", err)
	}
	if _, ok := scheduler.NextPiece(nil); !ok {
		t.Fatal("failed task should be assignable")
	}
	if err := scheduler.Complete(0); err != nil {
		t.Fatalf("Complete() failed: %v", err)
	}

	verified, total := scheduler.Progress()
	if verified != 1 || total != 1 {
		t.Fatalf("Progress() = %d/%d, want 1/1", verified, total)
	}
	if !scheduler.Done() {
		t.Fatal("Done() = false, want true")
	}
}

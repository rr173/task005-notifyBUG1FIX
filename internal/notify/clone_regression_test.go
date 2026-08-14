package notify

import (
	"testing"
	"time"
)

// TestMarkSent_ReturnedCopyIsIndependent 是一个固定的回归用例：
// MarkSent 返回给调用方的通知必须是独立副本，调用方修改返回结果中的
// SentAt 字段不能“穿透”到服务内部保存的数据。
//
// 运行命令（与基线、修复后完全一致）：
//
//	go test ./internal/notify/ -run TestMarkSent_ReturnedCopyIsIndependent -v
//
// 断言：篡改返回值的 SentAt 后，再次查询得到的内部 SentAt 仍等于标记时的原始时间。
func TestMarkSent_ReturnedCopyIsIndependent(t *testing.T) {
	s := New()

	created := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	if _, err := s.Create(CreateInput{
		ID:        "N1",
		Recipient: "user-a",
		Content:   "你好",
		Priority:  PriorityNormal,
	}, created); err != nil {
		t.Fatalf("Create 失败: %v", err)
	}

	sentAt := created.Add(time.Hour)
	returned, err := s.MarkSent("N1", sentAt)
	if err != nil {
		t.Fatalf("MarkSent 失败: %v", err)
	}
	if returned.SentAt == nil {
		t.Fatalf("返回值的 SentAt 为空")
	}
	// 记录标记已发送时服务内部应当保存的原始发送时间。
	wantSentAt := *returned.SentAt

	// 模拟调用方篡改返回结果中的发送时间。
	tampered := wantSentAt.Add(24 * time.Hour)
	*returned.SentAt = tampered

	// 重新查询内部存储，发送时间不应被外部篡改影响。
	got, err := s.Get("N1")
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if got.SentAt == nil {
		t.Fatalf("内部 SentAt 为空")
	}
	if !got.SentAt.Equal(wantSentAt) {
		t.Fatalf("内部 SentAt 被外部篡改: want %v, got %v", wantSentAt, got.SentAt)
	}
	if got.SentAt.Equal(tampered) {
		t.Fatalf("内部 SentAt 等于被篡改的值: got %v", got.SentAt)
	}

	// 顺便保证返回值本身确实反映了篡改（确认测试本身写对了）。
	if !returned.SentAt.Equal(tampered) {
		t.Fatalf("返回值未反映篡改: got %v, want %v", returned.SentAt, tampered)
	}
}

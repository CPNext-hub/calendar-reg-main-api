package externalapi

import (
	"context"
	"errors"
	"testing"
	"time"

	pb "github.com/CPNext-hub/calendar-reg-main-api/proto/gen/coursepb"
	"google.golang.org/grpc"
)

// ---- mock CourseServiceClient ----

type mockCourseServiceClient struct {
	resp *pb.FetchByCodeResponse
	err  error
}

func (m *mockCourseServiceClient) FetchByCode(
	ctx context.Context,
	in *pb.FetchByCodeRequest,
	opts ...grpc.CallOption,
) (*pb.FetchByCodeResponse, error) {
	return m.resp, m.err
}

// ---- FetchByCode tests ----

func TestFetchByCode_Success(t *testing.T) {
	mock := &mockCourseServiceClient{
		resp: &pb.FetchByCodeResponse{
			Code:         "CP353004",
			NameEn:       "Software Engineering",
			NameTh:       "วิศวกรรมซอฟต์แวร์",
			Faculty:      "วิทยาลัยการคอมพิวเตอร์",
			Department:   "วิทยาการคอมพิวเตอร์",
			Credits:      "3 (2-2-5)",
			Prerequisite: "CP353002",
			Semester:     2,
			Year:         2568,
			Sections: []*pb.Section{
				{
					Number:     "01",
					Seats:      40,
					Instructor: []string{"ผศ.ดร.ชิตสุธา สุ่มเล็ก"},
					ExamDate:   "31 มี.ค. 2569 เวลา 13:00 - 16:00",
					Note:       "หมายเหตุ",
					Campus:     "ขอนแก่น",
					Program:    "ปริญญาตรี ภาคปกติ",
					Schedules: []*pb.Schedule{
						{Day: "จันทร์", Time: "13:00-15:00", Room: "CP9 CP9127", Type: "C"},
					},
				},
			},
		},
	}

	api := &courseExternalAPI{client: mock}
	before := time.Now()

	course, err := api.FetchByCode(context.Background(), "CP353004", 2568, 2)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Code should be overwritten from the argument
	if course.Code != "CP353004" {
		t.Errorf("expected code CP353004, got %s", course.Code)
	}
	if course.NameEN != "Software Engineering" {
		t.Errorf("expected NameEN 'Software Engineering', got %s", course.NameEN)
	}
	if course.NameTH != "วิศวกรรมซอฟต์แวร์" {
		t.Errorf("expected NameTH, got %s", course.NameTH)
	}
	if course.Faculty != "วิทยาลัยการคอมพิวเตอร์" {
		t.Errorf("expected Faculty, got %s", course.Faculty)
	}
	if course.Department != "วิทยาการคอมพิวเตอร์" {
		t.Errorf("expected Department, got %s", course.Department)
	}
	if course.Credits != "3 (2-2-5)" {
		t.Errorf("expected Credits '3 (2-2-5)', got %s", course.Credits)
	}
	if course.Prerequisite != "CP353002" {
		t.Errorf("expected Prerequisite CP353002, got %s", course.Prerequisite)
	}
	if course.Semester != 2 {
		t.Errorf("expected Semester 2, got %d", course.Semester)
	}
	if course.Year != 2568 {
		t.Errorf("expected Year 2568, got %d", course.Year)
	}
	if course.UpdatedAt.Before(before) {
		t.Error("expected UpdatedAt to be set to now or later")
	}

	if len(course.Sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(course.Sections))
	}
	sec := course.Sections[0]
	if sec.Number != "01" {
		t.Errorf("expected section number '01', got %s", sec.Number)
	}
	if sec.Seats != 40 {
		t.Errorf("expected seats 40, got %d", sec.Seats)
	}
	if len(sec.Instructor) != 1 || sec.Instructor[0] != "ผศ.ดร.ชิตสุธา สุ่มเล็ก" {
		t.Errorf("expected instructor [ผศ.ดร.ชิตสุธา สุ่มเล็ก], got %v", sec.Instructor)
	}
	if sec.Campus != "ขอนแก่น" {
		t.Errorf("expected campus 'ขอนแก่น', got %s", sec.Campus)
	}
	if sec.Program != "ปริญญาตรี ภาคปกติ" {
		t.Errorf("expected program, got %s", sec.Program)
	}
	if sec.Note != "หมายเหตุ" {
		t.Errorf("expected note, got %s", sec.Note)
	}
	// Exam should be parsed
	if sec.ExamStart.IsZero() {
		t.Error("expected ExamStart to be parsed")
	}
	if sec.ExamEnd.IsZero() {
		t.Error("expected ExamEnd to be parsed")
	}

	if len(sec.Schedules) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(sec.Schedules))
	}
	sched := sec.Schedules[0]
	if sched.Day != "จันทร์" {
		t.Errorf("expected day 'จันทร์', got %s", sched.Day)
	}
	if sched.Room != "CP9 CP9127" {
		t.Errorf("expected room, got %s", sched.Room)
	}
	if sched.Type != "C" {
		t.Errorf("expected type 'C', got %s", sched.Type)
	}
	// StartTime should be 13:00
	if sched.StartTime.Hour() != 13 || sched.StartTime.Minute() != 0 {
		t.Errorf("expected start 13:00, got %v", sched.StartTime)
	}
	if sched.EndTime.Hour() != 15 || sched.EndTime.Minute() != 0 {
		t.Errorf("expected end 15:00, got %v", sched.EndTime)
	}
}

func TestFetchByCode_GRPCError(t *testing.T) {
	mock := &mockCourseServiceClient{
		err: errors.New("connection refused"),
	}
	api := &courseExternalAPI{client: mock}

	_, err := api.FetchByCode(context.Background(), "CS101", 2568, 1)
	if err == nil {
		t.Fatal("expected error from gRPC")
	}
	if err.Error() != "connection refused" {
		t.Errorf("expected 'connection refused', got %q", err.Error())
	}
}

// ---- protoToCourse edge-case tests ----

func TestProtoToCourse_InvalidScheduleTime(t *testing.T) {
	resp := &pb.FetchByCodeResponse{
		Sections: []*pb.Section{
			{
				Schedules: []*pb.Schedule{
					{Day: "จันทร์", Time: "invalid", Room: "R1", Type: "L"},
				},
			},
		},
	}

	course := protoToCourse(resp)
	if len(course.Sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(course.Sections))
	}
	sched := course.Sections[0].Schedules[0]
	// time should remain zero when parsing fails
	if !sched.StartTime.IsZero() {
		t.Error("expected zero StartTime for invalid time")
	}
	if !sched.EndTime.IsZero() {
		t.Error("expected zero EndTime for invalid time")
	}
}

func TestProtoToCourse_BadStartTime(t *testing.T) {
	resp := &pb.FetchByCodeResponse{
		Sections: []*pb.Section{
			{
				Schedules: []*pb.Schedule{
					{Day: "จันทร์", Time: "XX:XX-15:00", Room: "R1", Type: "L"},
				},
			},
		},
	}

	course := protoToCourse(resp)
	sched := course.Sections[0].Schedules[0]
	if !sched.StartTime.IsZero() {
		t.Error("expected zero StartTime for bad start time")
	}
}

func TestProtoToCourse_EmptyExamAndMidterm(t *testing.T) {
	resp := &pb.FetchByCodeResponse{
		Sections: []*pb.Section{
			{
				ExamDate:    "",
				MidtermDate: "",
			},
		},
	}

	course := protoToCourse(resp)
	sec := course.Sections[0]
	if !sec.ExamStart.IsZero() {
		t.Error("expected zero ExamStart for empty exam date")
	}
	if !sec.MidtermStart.IsZero() {
		t.Error("expected zero MidtermStart for empty midterm date")
	}
}

func TestProtoToCourse_InvalidExamDate(t *testing.T) {
	resp := &pb.FetchByCodeResponse{
		Sections: []*pb.Section{
			{
				ExamDate: "bad date string",
			},
		},
	}

	course := protoToCourse(resp)
	sec := course.Sections[0]
	// Should remain zero when parsing fails
	if !sec.ExamStart.IsZero() {
		t.Error("expected zero ExamStart for invalid exam date")
	}
}

func TestProtoToCourse_InvalidMidtermDate(t *testing.T) {
	resp := &pb.FetchByCodeResponse{
		Sections: []*pb.Section{
			{
				MidtermDate: "not a date",
			},
		},
	}

	course := protoToCourse(resp)
	sec := course.Sections[0]
	if !sec.MidtermStart.IsZero() {
		t.Error("expected zero MidtermStart for invalid midterm date")
	}
}

func TestProtoToCourse_ValidMidterm(t *testing.T) {
	resp := &pb.FetchByCodeResponse{
		Sections: []*pb.Section{
			{
				MidtermDate: "15 ก.พ. 2569 เวลา 09:00 - 12:00",
			},
		},
	}

	course := protoToCourse(resp)
	sec := course.Sections[0]
	if sec.MidtermStart.IsZero() {
		t.Error("expected MidtermStart to be parsed")
	}
	if sec.MidtermEnd.IsZero() {
		t.Error("expected MidtermEnd to be parsed")
	}
}

func TestProtoToCourse_ReservedFor(t *testing.T) {
	resp := &pb.FetchByCodeResponse{
		Sections: []*pb.Section{
			{
				ReservedFor: []string{"กลุ่ม A 50-49-1", "กลุ่ม B 50-49-2"},
			},
		},
	}

	course := protoToCourse(resp)
	sec := course.Sections[0]
	if len(sec.ReservedFor) != 2 {
		t.Errorf("expected 2 reserved entries, got %d", len(sec.ReservedFor))
	}
}

func TestProtoToCourse_NoSections(t *testing.T) {
	resp := &pb.FetchByCodeResponse{
		Code:   "CS101",
		NameEn: "Intro",
	}

	course := protoToCourse(resp)
	if len(course.Sections) != 0 {
		t.Errorf("expected 0 sections, got %d", len(course.Sections))
	}
	if course.NameEN != "Intro" {
		t.Errorf("expected NameEN 'Intro', got %s", course.NameEN)
	}
}

// ---- NewCourseExternalAPI test ----

func TestNewCourseExternalAPI(t *testing.T) {
	// Verify the constructor returns a non-nil value that implements the interface
	mock := &mockCourseServiceClient{}
	api := NewCourseExternalAPI(mockConn{mock: mock})
	if api == nil {
		t.Fatal("expected non-nil API")
	}
}

// mockConn implements grpc.ClientConnInterface to pass to NewCourseExternalAPI.
// In practice NewCourseServiceClient calls conn.Invoke which we already test via mockCourseServiceClient.
type mockConn struct {
	mock *mockCourseServiceClient
}

func (m mockConn) Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
	return nil
}

func (m mockConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
}

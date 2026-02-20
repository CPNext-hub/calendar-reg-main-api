package externalapi

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/repository"
	pb "github.com/CPNext-hub/calendar-reg-main-api/proto/gen/coursepb"
	"google.golang.org/grpc"
)

// courseExternalAPI is the gRPC implementation of repository.CourseExternalAPI.
type courseExternalAPI struct {
	client pb.CourseServiceClient
}

// NewCourseExternalAPI creates a new gRPC-based CourseExternalAPI.
func NewCourseExternalAPI(conn grpc.ClientConnInterface) repository.CourseExternalAPI {
	return &courseExternalAPI{client: pb.NewCourseServiceClient(conn)}
}

// FetchByCode calls the external gRPC service and returns a parsed Course entity.
func (a *courseExternalAPI) FetchByCode(ctx context.Context, code string, acadyear, semester int) (*entity.Course, error) {
	resp, err := a.client.FetchByCode(ctx, &pb.FetchByCodeRequest{
		Code:     code,
		Acadyear: int32(acadyear),
		Semester: int32(semester),
	})
	if err != nil {
		return nil, err
	}

	course := protoToCourse(resp)
	course.Code = code
	course.UpdatedAt = time.Now()
	return course, nil
}

// protoToCourse converts a gRPC FetchByCodeResponse to a domain Course entity.
func protoToCourse(resp *pb.FetchByCodeResponse) *entity.Course {
	sections := make([]entity.Section, len(resp.Sections))
	for i, s := range resp.Sections {
		schedules := make([]entity.Schedule, len(s.Schedules))
		for j, sc := range s.Schedules {
			var startTimeStr, endTimeStr string
			parts := strings.Split(sc.Time, "-")
			if len(parts) == 2 {
				startTimeStr = strings.TrimSpace(parts[0])
				endTimeStr = strings.TrimSpace(parts[1])
			}

			schedules[j] = entity.Schedule{
				Day:       sc.Day,
				StartTime: startTimeStr,
				EndTime:   endTimeStr,
				Room:      sc.Room,
				Type:      sc.Type,
			}
		}

		var examStartStr, examEndStr string
		if s.ExamDate != "" {
			es, ee, err := parseThaiExamDate(s.ExamDate)
			if err == nil {
				examStartStr = es
				examEndStr = ee
			} else {
				log.Printf("Error parsing exam date: %v", err)
			}
		}

		var midtermStartStr, midtermEndStr string
		if s.MidtermDate != "" {
			ms, me, err := parseThaiExamDate(s.MidtermDate)
			if err == nil {
				midtermStartStr = ms
				midtermEndStr = me
			} else {
				log.Printf("Error parsing midterm date: %v", err)
			}
		}

		sections[i] = entity.Section{
			Number:       s.Number,
			Schedules:    schedules,
			Seats:        int(s.Seats),
			Instructor:   s.Instructor,
			ExamStart:    examStartStr,
			ExamEnd:      examEndStr,
			MidtermStart: midtermStartStr,
			MidtermEnd:   midtermEndStr,
			Note:         s.Note,
			ReservedFor:  s.ReservedFor,
			Campus:       s.Campus,
			Program:      s.Program,
		}
	}

	return &entity.Course{
		Code:         resp.Code,
		NameEN:       resp.NameEn,
		NameTH:       resp.NameTh,
		Faculty:      resp.Faculty,
		Department:   resp.Department,
		Credits:      resp.Credits,
		Prerequisite: resp.Prerequisite,
		Semester:     int(resp.Semester),
		Year:         int(resp.Year),
		Sections:     sections,
	}
}

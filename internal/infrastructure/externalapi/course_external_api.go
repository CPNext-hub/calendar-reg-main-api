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
func (a *courseExternalAPI) FetchByCode(ctx context.Context, code string) (*entity.Course, error) {
	resp, err := a.client.FetchByCode(ctx, &pb.FetchByCodeRequest{Code: code})
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
			var startTime, endTime time.Time
			parts := strings.Split(sc.Time, "-")
			if len(parts) == 2 {
				st, err1 := time.Parse("15:04", strings.TrimSpace(parts[0]))
				et, err2 := time.Parse("15:04", strings.TrimSpace(parts[1]))
				if err1 == nil && err2 == nil {
					startTime = st
					endTime = et
				} else {
					log.Printf("Error parsing time for schedule: %v, %v", err1, err2)
				}
			}

			schedules[j] = entity.Schedule{
				Day:       sc.Day,
				StartTime: startTime,
				EndTime:   endTime,
				Room:      sc.Room,
				Type:      sc.Type,
			}
		}

		var examStart, examEnd time.Time
		if s.ExamDate != "" {
			es, ee, err := parseThaiExamDate(s.ExamDate)
			if err == nil {
				examStart = es
				examEnd = ee
			} else {
				log.Printf("Error parsing exam date: %v", err)
			}
		}

		sections[i] = entity.Section{
			Number:     s.Number,
			Schedules:  schedules,
			Seats:      int(s.Seats),
			Instructor: s.Instructor,
			ExamStart:  examStart,
			ExamEnd:    examEnd,
		}
	}

	return &entity.Course{
		Code:         resp.Code,
		NameEN:       resp.NameEn,
		NameTH:       resp.NameTh,
		Faculty:      resp.Faculty,
		Credits:      resp.Credits,
		Prerequisite: resp.Prerequisite,
		Semester:     int(resp.Semester),
		Year:         int(resp.Year),
		Program:      resp.Program,
		Sections:     sections,
	}
}

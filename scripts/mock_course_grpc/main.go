// Mock gRPC server for the CourseService.
// Usage: go run scripts/mock_course_grpc/main.go
package main

import (
	"context"
	"log"
	"net"
	"time"

	pb "github.com/CPNext-hub/calendar-reg-main-api/proto/gen/coursepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var courses = map[string]*pb.FetchByCodeResponse{
	"CP353004": {
		Code:         "CP353004",
		NameEn:       "Software Engineering",
		NameTh:       "วิศวกรรมซอฟต์แวร์",
		Faculty:      "College of Computing",
		Credits:      "3(2-2-5)",
		Prerequisite: "CP353002",
		Semester:     1,
		Year:         2567,
		Sections: []*pb.Section{
			{
				Number:      "02",
				Seats:       40,
				Instructor:  "Assoc. Prof. Dr. Chitsutha Soomlek",
				ExamDate:    "31 มี.ค. 2567 เวลา 13:00 - 16:00",
				MidtermDate: "15 ก.พ. 2567 เวลา 09:00 - 12:00",
				Note:        "",
				ReservedFor: []string{},
				Campus:      "ขอนแก่น",
				Program:     "Undergraduate (Regular)",
				Schedules: []*pb.Schedule{
					{Day: "Monday", Time: "15:00-17:00", Room: "CP9127", Type: "Lecture"},
					{Day: "Wednesday", Time: "13:00-15:00", Room: "CP9127", Type: "Lab"},
				},
			},
		},
	},
	"CP353002": {
		Code:         "CP353002",
		NameEn:       "Object-Oriented Programming",
		NameTh:       "การเขียนโปรแกรมเชิงวัตถุ",
		Faculty:      "College of Computing",
		Credits:      "3(2-2-5)",
		Prerequisite: "",
		Semester:     1,
		Year:         2567,
		Sections: []*pb.Section{
			{
				Number:      "01",
				Seats:       60,
				Instructor:  "Dr. Somchai Prasit",
				ExamDate:    "28 มี.ค. 2567 เวลา 09:00 - 12:00",
				MidtermDate: "10 ก.พ. 2567 เวลา 09:00 - 12:00",
				Note:        "ผู้สอบไม่ผ่าน",
				ReservedFor: []string{"ผู้ที่สอบไม่ผ่าน50-49-1"},
				Campus:      "ขอนแก่น",
				Program:     "Undergraduate (Regular)",
				Schedules: []*pb.Schedule{
					{Day: "Tuesday", Time: "09:00-11:00", Room: "CP9101", Type: "Lecture"},
					{Day: "Thursday", Time: "13:00-15:00", Room: "CP9103", Type: "Lab"},
				},
			},
		},
	},
	"CP353006": {
		Code:         "CP353006",
		NameEn:       "Database Systems",
		NameTh:       "ระบบฐานข้อมูล",
		Faculty:      "College of Computing",
		Credits:      "3(2-2-5)",
		Prerequisite: "CP353002",
		Semester:     2,
		Year:         2567,
		Sections: []*pb.Section{
			{
				Number:      "01",
				Seats:       45,
				Instructor:  "Asst. Prof. Dr. Wanida Kanarkard",
				ExamDate:    "30 มี.ค. 2567 เวลา 09:00 - 12:00",
				MidtermDate: "12 ก.พ. 2567 เวลา 13:00 - 16:00",
				Note:        "Closed",
				ReservedFor: []string{},
				Campus:      "หนองคาย",
				Program:     "Undergraduate (Regular)",
				Schedules: []*pb.Schedule{
					{Day: "Monday", Time: "09:00-11:00", Room: "CP9205", Type: "Lecture"},
					{Day: "Friday", Time: "13:00-15:00", Room: "CP9205", Type: "Lab"},
				},
			},
		},
	},
}

type server struct {
	pb.UnimplementedCourseServiceServer
}

func (s *server) FetchByCode(_ context.Context, req *pb.FetchByCodeRequest) (*pb.FetchByCodeResponse, error) {
	log.Printf("FetchByCode request: code=%s acadyear=%d semester=%d", req.Code, req.Acadyear, req.Semester)

	// Simulate slow response
	time.Sleep(20 * time.Second)

	course, ok := courses[req.Code]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "course '%s' not found", req.Code)
	}
	return course, nil
}

func main() {
	const port = ":50051"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterCourseServiceServer(s, &server{})

	log.Printf("🚀 Mock Course gRPC server running at %s", port)
	log.Printf("📚 Available courses: CP353004, CP353002, CP353006")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

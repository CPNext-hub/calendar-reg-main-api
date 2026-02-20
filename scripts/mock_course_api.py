"""Mock Course API — returns sample course JSON for testing.

Usage:
  GET /<code>  → returns course data if found, 404 otherwise.
  Example: GET /CP353004
"""

from http.server import HTTPServer, BaseHTTPRequestHandler
import json
import time

COURSES = {
    "CP353004": {
        "code": "CP353004",
        "name_en": "Software Engineering",
        "name_th": "วิศวกรรมซอฟต์แวร์",
        "faculty": "College of Computing",
        "credits": "3(2-2-5)",
        "prerequisite": "CP353002",
        "semester": 1,
        "year": 2567,
        "program": "Undergraduate (Regular)",
        "sections": [
            {
                "number": "02",
                "seats": 40,
                "instructor": ["Assoc. Prof. Dr. Chitsutha Soomlek"],
                "exam_date": "31 มี.ค. 2567 เวลา 13:00 - 16:00",
                "schedules": [
                    {"day": "Monday", "time": "15:00-17:00", "room": "CP9127", "type": "Lecture"},
                    {"day": "Wednesday", "time": "13:00-15:00", "room": "CP9127", "type": "Lab"},
                ],
            }
        ],
    },
    "CP353002": {
        "code": "CP353002",
        "name_en": "Object-Oriented Programming",
        "name_th": "การเขียนโปรแกรมเชิงวัตถุ",
        "faculty": "College of Computing",
        "credits": "3(2-2-5)",
        "prerequisite": "",
        "semester": 1,
        "year": 2567,
        "program": "Undergraduate (Regular)",
        "sections": [
            {
                "number": "01",
                "seats": 60,
                "instructor": ["Dr. Somchai Prasit"],
                "exam_date": "28 มี.ค. 2567 เวลา 09:00 - 12:00",
                "schedules": [
                    {"day": "Tuesday", "time": "09:00-11:00", "room": "CP9101", "type": "Lecture"},
                    {"day": "Thursday", "time": "13:00-15:00", "room": "CP9103", "type": "Lab"},
                ],
            }
        ],
    },
    "CP353006": {
        "code": "CP353006",
        "name_en": "Database Systems",
        "name_th": "ระบบฐานข้อมูล",
        "faculty": "College of Computing",
        "credits": "3(2-2-5)",
        "prerequisite": "CP353002",
        "semester": 2,
        "year": 2567,
        "program": "Undergraduate (Regular)",
        "sections": [
            {
                "number": "01",
                "seats": 45,
                "instructor": ["Asst. Prof. Dr. Wanida Kanarkard"],
                "exam_date": "30 มี.ค. 2567 เวลา 09:00 - 12:00",
                "schedules": [
                    {"day": "Monday", "time": "09:00-11:00", "room": "CP9205", "type": "Lecture"},
                    {"day": "Friday", "time": "13:00-15:00", "room": "CP9205", "type": "Lab"},
                ],
            }
        ],
    },
}


class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        time.sleep(20)  # Simulate slow response

        # Extract course code from path: /CP353004 → "CP353004"
        code = self.path.strip("/")

        if not code:
            # No code provided — list all available codes
            try:
                self.send_response(200)
                self.send_header("Content-Type", "application/json; charset=utf-8")
                self.end_headers()
                body = {"available_codes": list(COURSES.keys())}
                self.wfile.write(json.dumps(body, ensure_ascii=False, indent=2).encode("utf-8"))
            except BrokenPipeError:
                print("⚠️  Client disconnected (timeout)")
            return

        course = COURSES.get(code)
        if course is None:
            try:
                self.send_response(404)
                self.send_header("Content-Type", "application/json; charset=utf-8")
                self.end_headers()
                body = {"error": f"course '{code}' not found"}
                self.wfile.write(json.dumps(body, ensure_ascii=False, indent=2).encode("utf-8"))
            except BrokenPipeError:
                print("⚠️  Client disconnected (timeout)")
            return

        try:
            self.send_response(200)
            self.send_header("Content-Type", "application/json; charset=utf-8")
            self.end_headers()
            self.wfile.write(json.dumps(course, ensure_ascii=False, indent=2).encode("utf-8"))
        except BrokenPipeError:
            print("⚠️  Client disconnected (timeout)")


if __name__ == "__main__":
    PORT = 8888
    server = HTTPServer(("0.0.0.0", PORT), Handler)
    print(f"🚀 Mock Course API running at http://localhost:{PORT}")
    print(f"📚 Available courses: {', '.join(COURSES.keys())}")
    server.serve_forever()

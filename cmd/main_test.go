package main

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRoot(t *testing.T) {
	// Setup the app
	app := setupApp()

	// Create a new request
	req := httptest.NewRequest("GET", "/", nil)

	// Perform the request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	// Check status code
	assert.Equal(t, 200, resp.StatusCode, "Status code should be 200")

	// Check response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Check if body equals "Hello World"
	assert.Equal(t, "Hello World", string(body), "Response body should be 'Hello World'")
}

// เพิ่ม test case สำหรับกรณีที่พบข้อมูลลูกค้า
func TestGetCustomerDetailsSuccess(t *testing.T) {
	// Setup the app
	app := setupApp()

	// สร้าง request สำหรับลูกค้าที่มีอยู่ในระบบ
	req := httptest.NewRequest("GET", "/staff/customers/12345", nil)

	// ทดสอบการเรียก API
	resp, err := app.Test(req)
	assert.NoError(t, err, "ไม่ควรมีข้อผิดพลาดในการส่ง request")

	// ตรวจสอบ status code
	assert.Equal(t, 200, resp.StatusCode, "Status code ควรเป็น 200")

	// ตรวจสอบ response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "ไม่ควรมีข้อผิดพลาดในการอ่าน response body")

	// แปลง JSON response เป็น map
	var customer map[string]interface{}
	err = json.Unmarshal(body, &customer)
	assert.NoError(t, err, "Response ควรเป็น JSON ที่ถูกต้อง")

	// ตรวจสอบข้อมูลลูกค้า
	assert.Equal(t, "12345", customer["id"], "ID ลูกค้าควรตรงกัน")
	assert.Equal(t, "สมชาย", customer["first_name"], "ชื่อลูกค้าควรตรงกัน")
	assert.Equal(t, "ใจดี", customer["last_name"], "นามสกุลลูกค้าควรตรงกัน")
}

// เพิ่ม test case สำหรับกรณีที่ไม่พบข้อมูลลูกค้า
func TestGetCustomerDetailsNotFound(t *testing.T) {
	// Setup the app
	app := setupApp()

	// สร้าง request สำหรับลูกค้าที่ไม่มีในระบบ
	req := httptest.NewRequest("GET", "/staff/customers/99999", nil)

	// ทดสอบการเรียก API
	resp, err := app.Test(req)
	assert.NoError(t, err, "ไม่ควรมีข้อผิดพลาดในการส่ง request")

	// ตรวจสอบ status code
	assert.Equal(t, 404, resp.StatusCode, "Status code ควรเป็น 404 สำหรับกรณีไม่พบข้อมูล")

	// ตรวจสอบ error response
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "ไม่ควรมีข้อผิดพลาดในการอ่าน response body")

	var errorResponse map[string]interface{}
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err, "Response ควรเป็น JSON ที่ถูกต้อง")
	assert.Contains(t, errorResponse, "error", "Response ควรมีข้อความแจ้ง error")
}

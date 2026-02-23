package email

import "fmt"

// LeaveRequestAssignedTemplate generates HTML for leave request assignment notification
func LeaveRequestAssignedTemplate(requesterName, leaveType string, days int, startDate, endDate string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
        }
        .header {
            background-color: #4CAF50;
            color: white;
            padding: 20px;
            text-align: center;
        }
        .content {
            padding: 20px;
            background-color: #f9f9f9;
        }
        .action-box {
            background-color: white;
            border-left: 4px solid #4CAF50;
            padding: 15px;
            margin: 20px 0;
        }
        .detail-row {
            margin: 10px 0;
        }
        .label {
            font-weight: bold;
            color: #555;
        }
        .value {
            color: #333;
        }
        .footer {
            padding: 20px;
            text-align: center;
            font-size: 12px;
            color: #777;
        }
        .button {
            display: inline-block;
            padding: 12px 24px;
            background-color: #4CAF50;
            color: white;
            text-decoration: none;
            border-radius: 4px;
            margin-top: 15px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h2>Leave Request Assigned for Review</h2>
    </div>
    <div class="content">
        <p>Hello,</p>
        <p>A leave request has been assigned to you for review.</p>

        <div class="action-box">
            <div class="detail-row">
                <span class="label">Employee:</span>
                <span class="value">%s</span>
            </div>
            <div class="detail-row">
                <span class="label">Leave Type:</span>
                <span class="value">%s</span>
            </div>
            <div class="detail-row">
                <span class="label">Duration:</span>
                <span class="value">%d days</span>
            </div>
            <div class="detail-row">
                <span class="label">Period:</span>
                <span class="value">%s to %s</span>
            </div>
        </div>

        <p>Please log in to the HR System to review and take action on this request.</p>
    </div>
    <div class="footer">
        <p>This is an automated message from the HR System. Please do not reply to this email.</p>
    </div>
</body>
</html>
`, requesterName, leaveType, days, startDate, endDate)
}

// LeaveRequestApprovedTemplate generates HTML for leave request approval notification
func LeaveRequestApprovedTemplate(employeeName, leaveType string, days int, startDate, endDate, approverName string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
        }
        .header {
            background-color: #4CAF50;
            color: white;
            padding: 20px;
            text-align: center;
        }
        .content {
            padding: 20px;
            background-color: #f9f9f9;
        }
        .success-box {
            background-color: #d4edda;
            border-left: 4px solid #28a745;
            padding: 15px;
            margin: 20px 0;
        }
        .detail-row {
            margin: 10px 0;
        }
        .label {
            font-weight: bold;
            color: #555;
        }
        .value {
            color: #333;
        }
        .footer {
            padding: 20px;
            text-align: center;
            font-size: 12px;
            color: #777;
        }
    </style>
</head>
<body>
    <div class="header">
        <h2>✓ Leave Request Approved</h2>
    </div>
    <div class="content">
        <p>Hello %s,</p>
        <p>Great news! Your leave request has been approved.</p>

        <div class="success-box">
            <div class="detail-row">
                <span class="label">Leave Type:</span>
                <span class="value">%s</span>
            </div>
            <div class="detail-row">
                <span class="label">Duration:</span>
                <span class="value">%d days</span>
            </div>
            <div class="detail-row">
                <span class="label">Period:</span>
                <span class="value">%s to %s</span>
            </div>
            <div class="detail-row">
                <span class="label">Approved by:</span>
                <span class="value">%s</span>
            </div>
        </div>

        <p>Your leave has been officially approved. Enjoy your time off!</p>
    </div>
    <div class="footer">
        <p>This is an automated message from the HR System. Please do not reply to this email.</p>
    </div>
</body>
</html>
`, employeeName, leaveType, days, startDate, endDate, approverName)
}

// LeaveRequestRejectedTemplate generates HTML for leave request rejection notification
func LeaveRequestRejectedTemplate(employeeName, leaveType string, days int, startDate, endDate, reviewerName, reason string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
        }
        .header {
            background-color: #f44336;
            color: white;
            padding: 20px;
            text-align: center;
        }
        .content {
            padding: 20px;
            background-color: #f9f9f9;
        }
        .warning-box {
            background-color: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 15px;
            margin: 20px 0;
        }
        .detail-row {
            margin: 10px 0;
        }
        .label {
            font-weight: bold;
            color: #555;
        }
        .value {
            color: #333;
        }
        .footer {
            padding: 20px;
            text-align: center;
            font-size: 12px;
            color: #777;
        }
    </style>
</head>
<body>
    <div class="header">
        <h2>Leave Request Not Approved</h2>
    </div>
    <div class="content">
        <p>Hello %s,</p>
        <p>We regret to inform you that your leave request has not been approved.</p>

        <div class="warning-box">
            <div class="detail-row">
                <span class="label">Leave Type:</span>
                <span class="value">%s</span>
            </div>
            <div class="detail-row">
                <span class="label">Duration:</span>
                <span class="value">%d days</span>
            </div>
            <div class="detail-row">
                <span class="label">Period:</span>
                <span class="value">%s to %s</span>
            </div>
            <div class="detail-row">
                <span class="label">Reviewed by:</span>
                <span class="value">%s</span>
            </div>
            %s
        </div>

        <p>If you have any questions, please contact your manager or HR department.</p>
    </div>
    <div class="footer">
        <p>This is an automated message from the HR System. Please do not reply to this email.</p>
    </div>
</body>
</html>
`, employeeName, leaveType, days, startDate, endDate, reviewerName, formatReason(reason))
}

func formatReason(reason string) string {
	if reason != "" {
		return fmt.Sprintf(`
            <div class="detail-row">
                <span class="label">Reason:</span>
                <span class="value">%s</span>
            </div>`, reason)
	}
	return ""
}

// GenericTaskAssignedTemplate generates HTML for generic task assignment notification
func GenericTaskAssignedTemplate(recipientName, taskName, taskDescription string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
        }
        .header {
            background-color: #4CAF50;
            color: white;
            padding: 20px;
            text-align: center;
        }
        .content {
            padding: 20px;
            background-color: #f9f9f9;
        }
        .task-box {
            background-color: white;
            border-left: 4px solid #4CAF50;
            padding: 15px;
            margin: 20px 0;
        }
        .detail-row {
            margin: 10px 0;
        }
        .label {
            font-weight: bold;
            color: #555;
        }
        .value {
            color: #333;
        }
        .footer {
            padding: 20px;
            text-align: center;
            font-size: 12px;
            color: #777;
        }
    </style>
</head>
<body>
    <div class="header">
        <h2>New Task Assigned</h2>
    </div>
    <div class="content">
        <p>Hello %s,</p>
        <p>A new task has been assigned to you for your action.</p>

        <div class="task-box">
            <div class="detail-row">
                <span class="label">Task:</span>
                <span class="value">%s</span>
            </div>
            <div class="detail-row">
                <span class="label">Description:</span>
                <span class="value">%s</span>
            </div>
        </div>

        <p>Please log in to the HR System to review and take action on this task.</p>
    </div>
    <div class="footer">
        <p>This is an automated message from the HR System. Please do not reply to this email.</p>
    </div>
</body>
</html>
`, recipientName, taskName, taskDescription)
}

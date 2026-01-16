import { APIRequestContext } from '@playwright/test';

/**
 * API Response wrapper
 */
export interface ApiResponse<T> {
  data: T;
}

/**
 * API Error response
 */
export interface ApiError {
  error: {
    code: string;
    message: string;
  };
}

/**
 * Test credentials for seeded admin user
 */
export const TEST_CREDENTIALS = {
  email: 'admin1@example.com',
  password: 'password123',
};

/**
 * API endpoints
 */
export const ENDPOINTS = {
  // Auth
  login: '/api/v1/auth/login',
  setup: '/api/v1/setup',
  registerByInvite: '/api/v1/auth/register-by-invite',
  passwordResetStatus: '/api/v1/auth/password-reset-status',
  resetPassword: '/api/v1/auth/reset-password',

  // Health
  health: '/health',

  // Admin
  adminChangePassword: '/api/v1/admins/me/change-password',
  adminAllowPasswordReset: (adminId: string) => `/api/v1/admins/${adminId}/allow-password-reset`,

  // Tenant
  tenant: '/api/v1/tenants/me',
  managerPermissions: '/api/v1/settings/manager-permissions',

  // Invitations
  invitations: '/api/v1/invitations',
  acceptInvitation: (token: string) => `/api/v1/invitations/accept/${token}`,

  // Members
  members: '/api/v1/members',
  member: (id: string) => `/api/v1/members/${id}`,
  membersMe: '/api/v1/members/me',
  membersRecentAttendance: '/api/v1/members/recent-attendance',
  membersBulkImport: '/api/v1/members/bulk-import',
  membersBulkUpdateRoles: '/api/v1/members/bulk-update-roles',

  // Roles
  roles: '/api/v1/roles',
  role: (id: string) => `/api/v1/roles/${id}`,

  // Events
  events: '/api/v1/events',
  event: (id: string) => `/api/v1/events/${id}`,
  eventGenerateBusinessDays: (eventId: string) => `/api/v1/events/${eventId}/generate-business-days`,
  eventGroups: (eventId: string) => `/api/v1/events/${eventId}/groups`,

  // Business Days
  businessDays: '/api/v1/business-days',
  businessDay: (id: string) => `/api/v1/business-days/${id}`,
  businessDaysByEvent: (eventId: string) => `/api/v1/events/${eventId}/business-days`,
  businessDayApplyTemplate: (id: string) => `/api/v1/business-days/${id}/apply-template`,
  businessDaySaveAsTemplate: (id: string) => `/api/v1/business-days/${id}/save-as-template`,

  // Shift Slots
  shiftSlots: '/api/v1/shift-slots',
  shiftSlot: (id: string) => `/api/v1/shift-slots/${id}`,
  shiftSlotsByBusinessDay: (businessDayId: string) => `/api/v1/business-days/${businessDayId}/shift-slots`,

  // Shift Assignments
  shiftAssignments: '/api/v1/shift-assignments',
  shiftAssignment: (id: string) => `/api/v1/shift-assignments/${id}`,
  shiftAssignmentStatus: (id: string) => `/api/v1/shift-assignments/${id}/status`,

  // Templates
  templates: (eventId: string) => `/api/v1/events/${eventId}/templates`,
  template: (eventId: string, templateId: string) => `/api/v1/events/${eventId}/templates/${templateId}`,

  // Instances
  instances: (eventId: string) => `/api/v1/events/${eventId}/instances`,
  instance: (id: string) => `/api/v1/instances/${id}`,

  // Member Groups
  memberGroups: '/api/v1/member-groups',
  memberGroup: (id: string) => `/api/v1/member-groups/${id}`,
  memberGroupMembers: (id: string) => `/api/v1/member-groups/${id}/members`,

  // Role Groups
  roleGroups: '/api/v1/role-groups',
  roleGroup: (id: string) => `/api/v1/role-groups/${id}`,
  roleGroupRoles: (id: string) => `/api/v1/role-groups/${id}/roles`,

  // Attendance
  attendanceCollections: '/api/v1/attendance/collections',
  attendanceCollection: (id: string) => `/api/v1/attendance/collections/${id}`,
  attendanceCollectionClose: (id: string) => `/api/v1/attendance/collections/${id}/close`,
  attendanceCollectionResponses: (id: string) => `/api/v1/attendance/collections/${id}/responses`,

  // Schedules
  schedules: '/api/v1/schedules',
  schedule: (id: string) => `/api/v1/schedules/${id}`,
  scheduleDecide: (id: string) => `/api/v1/schedules/${id}/decide`,
  scheduleClose: (id: string) => `/api/v1/schedules/${id}/close`,
  scheduleResponses: (id: string) => `/api/v1/schedules/${id}/responses`,

  // Import
  imports: '/api/v1/imports',
  importMembers: '/api/v1/imports/members',
  importStatus: (id: string) => `/api/v1/imports/${id}/status`,
  importResult: (id: string) => `/api/v1/imports/${id}/result`,

  // Announcements
  announcements: '/api/v1/announcements',
  announcementsUnreadCount: '/api/v1/announcements/unread-count',
  announcementRead: (id: string) => `/api/v1/announcements/${id}/read`,
  announcementsReadAll: '/api/v1/announcements/read-all',

  // Tutorials
  tutorials: '/api/v1/tutorials',
  tutorial: (id: string) => `/api/v1/tutorials/${id}`,

  // Actual Attendance
  actualAttendance: '/api/v1/actual-attendance',

  // Public APIs
  publicAttendance: (token: string) => `/api/v1/public/attendance/${token}`,
  publicAttendanceResponses: (token: string) => `/api/v1/public/attendance/${token}/responses`,
  publicAttendanceMemberResponses: (token: string, memberId: string) => `/api/v1/public/attendance/${token}/members/${memberId}/responses`,
  publicMembers: '/api/v1/public/members',
  publicSchedule: (token: string) => `/api/v1/public/schedules/${token}`,
  publicScheduleResponses: (token: string) => `/api/v1/public/schedules/${token}/responses`,
  publicLicenseClaim: '/api/v1/public/license/claim',
};

/**
 * Helper to make authenticated API requests
 */
export class ApiClient {
  private token: string | null = null;

  constructor(private request: APIRequestContext) {}

  /**
   * Set authentication token
   */
  setToken(token: string) {
    this.token = token;
  }

  /**
   * Clear authentication token
   */
  clearToken() {
    this.token = null;
  }

  /**
   * Get headers with optional authentication
   */
  private getHeaders(): Record<string, string> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }
    return headers;
  }

  /**
   * GET request
   */
  async get<T>(endpoint: string, options?: { params?: Record<string, string> }): Promise<T> {
    let url = endpoint;
    if (options?.params) {
      const searchParams = new URLSearchParams(options.params);
      url = `${endpoint}?${searchParams.toString()}`;
    }

    const response = await this.request.get(url, {
      headers: this.getHeaders(),
    });

    return response.json();
  }

  /**
   * POST request
   */
  async post<T>(endpoint: string, data?: unknown): Promise<T> {
    const response = await this.request.post(endpoint, {
      headers: this.getHeaders(),
      data,
    });

    return response.json();
  }

  /**
   * PUT request
   */
  async put<T>(endpoint: string, data?: unknown): Promise<T> {
    const response = await this.request.put(endpoint, {
      headers: this.getHeaders(),
      data,
    });

    return response.json();
  }

  /**
   * PATCH request
   */
  async patch<T>(endpoint: string, data?: unknown): Promise<T> {
    const response = await this.request.patch(endpoint, {
      headers: this.getHeaders(),
      data,
    });

    return response.json();
  }

  /**
   * DELETE request
   */
  async delete<T>(endpoint: string): Promise<T> {
    const response = await this.request.delete(endpoint, {
      headers: this.getHeaders(),
    });

    return response.json();
  }

  /**
   * Raw request (for checking status codes)
   */
  async raw(method: string, endpoint: string, data?: unknown) {
    const options: { headers: Record<string, string>; data?: unknown } = {
      headers: this.getHeaders(),
    };
    if (data) {
      options.data = data;
    }

    switch (method.toUpperCase()) {
      case 'GET':
        return this.request.get(endpoint, options);
      case 'POST':
        return this.request.post(endpoint, options);
      case 'PUT':
        return this.request.put(endpoint, options);
      case 'PATCH':
        return this.request.patch(endpoint, options);
      case 'DELETE':
        return this.request.delete(endpoint, options);
      default:
        throw new Error(`Unsupported method: ${method}`);
    }
  }
}

// API Client for SmashQueue Backend

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

// Types
export interface User {
  id: number;
  username: string;
  name: string;
  bio: string;
  role: 'player' | 'organizer' | 'admin';
  created_at: string;
  updated_at: string;
}

export interface UserStats {
  user_id: number;
  total_matches: number;
  wins: number;
  losses: number;
  win_rate: number;
  current_streak: number;
  skill_level: string;
}

export interface UserProfile {
  user: User;
  stats: UserStats;
}

export interface AuthResponse {
  access_token: string;
  user: User;
}

export interface APIResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: number;
    message: string;
    details?: string;
  };
}

export interface QueueInfo {
  total_in_queue: number;
  your_position?: number;
  estimated_wait?: string;
  next_court?: string;
  currently_playing: QueueEntry[];
}

export interface QueueEntry {
  id: number;
  user_id: number;
  position: number;
  status: string;
  joined_at: string;
}

export interface GameScore {
  game: number;
  team1_score: number;
  team2_score: number;
}

export interface Match {
  id: number;
  court: string;
  team1: number[];
  team2: number[];
  scores: GameScore[];
  result: string;
  started_at: string;
  ended_at?: string;
}

export interface MatchHistory {
  match: Match;
  won: boolean;
}

// Token management
let accessToken: string | null = null;

export function setAccessToken(token: string | null) {
  accessToken = token;
  if (typeof window !== 'undefined') {
    if (token) {
      localStorage.setItem('access_token', token);
    } else {
      localStorage.removeItem('access_token');
    }
  }
}

export function getAccessToken(): string | null {
  if (!accessToken && typeof window !== 'undefined') {
    accessToken = localStorage.getItem('access_token');
  }
  return accessToken;
}

export function clearAuth() {
  accessToken = null;
  if (typeof window !== 'undefined') {
    localStorage.removeItem('access_token');
    localStorage.removeItem('user');
  }
}

export function setUser(user: User | null) {
  if (typeof window !== 'undefined') {
    if (user) {
      localStorage.setItem('user', JSON.stringify(user));
    } else {
      localStorage.removeItem('user');
    }
  }
}

export function getUser(): User | null {
  if (typeof window !== 'undefined') {
    const stored = localStorage.getItem('user');
    if (stored) {
      try {
        return JSON.parse(stored);
      } catch {
        return null;
      }
    }
  }
  return null;
}

export function isAuthenticated(): boolean {
  return getAccessToken() !== null;
}

export function isAdmin(): boolean {
  const user = getUser();
  return user?.role === 'admin';
}

export function isOrganizer(): boolean {
  const user = getUser();
  return user?.role === 'organizer' || user?.role === 'admin';
}

// API request helper
async function request<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<APIResponse<T>> {
  const token = getAccessToken();
  
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...options.headers,
  };

  if (token) {
    (headers as Record<string, string>)['Authorization'] = `Bearer ${token}`;
  }

  try {
    const response = await fetch(`${API_URL}${endpoint}`, {
      ...options,
      headers,
      credentials: 'include',
    });

    const data = await response.json();
    return data;
  } catch (error) {
    return {
      success: false,
      error: {
        code: 500,
        message: 'Network error',
        details: error instanceof Error ? error.message : 'Unknown error',
      },
    };
  }
}

// Auth API
export async function login(username: string, password: string): Promise<APIResponse<AuthResponse>> {
  const response = await request<AuthResponse>('/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });

  if (response.success && response.data) {
    setAccessToken(response.data.access_token);
    setUser(response.data.user);
  }

  return response;
}

export async function register(
  username: string,
  password: string,
  confirmPassword: string,
  phone?: string,
  name?: string
): Promise<APIResponse<AuthResponse>> {
  const response = await request<AuthResponse>('/register', {
    method: 'POST',
    body: JSON.stringify({
      username,
      password,
      confirm_password: confirmPassword,
      phone: phone || '',
      name: name || username,
    }),
  });

  if (response.success && response.data) {
    setAccessToken(response.data.access_token);
    setUser(response.data.user);
  }

  return response;
}

export async function logout(): Promise<void> {
  await request('/logout', { method: 'POST' });
  clearAuth();
}

// Profile API
export async function getProfile(): Promise<APIResponse<UserProfile>> {
  return request<UserProfile>('/profile');
}

export async function updateProfile(name: string, bio: string): Promise<APIResponse<User>> {
  return request<User>('/profile', {
    method: 'PUT',
    body: JSON.stringify({ name, bio }),
  });
}

// Queue API
export async function getQueueStatus(): Promise<APIResponse<QueueInfo>> {
  return request<QueueInfo>('/queue');
}

export async function joinQueue(): Promise<APIResponse<{ entry: QueueEntry; info: QueueInfo; message: string }>> {
  return request('/queue/join', { method: 'POST' });
}

export async function leaveQueue(): Promise<APIResponse<{ info: QueueInfo; message: string }>> {
  return request('/queue/leave', { method: 'POST' });
}

export async function callNextPlayers(): Promise<APIResponse<{ called: QueueEntry[]; count: number }>> {
  return request('/queue/call', { method: 'POST' });
}

// Match API
export async function getMatchHistory(): Promise<APIResponse<MatchHistory[]>> {
  return request<MatchHistory[]>('/matches');
}

export async function getActiveMatches(): Promise<APIResponse<Match[]>> {
  return request<Match[]>('/matches/active');
}

export async function createMatch(court: string, team1: number[], team2: number[]): Promise<APIResponse<Match>> {
  return request<Match>('/matches', {
    method: 'POST',
    body: JSON.stringify({ court, team1, team2 }),
  });
}

export async function recordMatchResult(matchId: number, scores: GameScore[]): Promise<APIResponse<Match>> {
  return request<Match>('/matches/result', {
    method: 'PUT',
    body: JSON.stringify({ match_id: matchId, scores }),
  });
}

// Admin API
export interface UserListItem {
  id: number;
  username: string;
  name: string;
  phone: string;
  role: string;
  hand_preference?: string;
  skill_tier?: string;
  is_active: boolean;
  skill_level: string;
  win_rate: number;
  total_matches?: number;
  wins?: number;
}

export async function getAllUsers(): Promise<APIResponse<UserListItem[]>> {
  return request<UserListItem[]>('/admin/users');
}

export async function updatePlayerAdmin(
  userId: number,
  handPreference: string,
  skillTier: string
): Promise<APIResponse<{ message: string }>> {
  return request(`/admin/users/${userId}`, {
    method: 'PUT',
    body: JSON.stringify({
      hand_preference: handPreference,
      skill_tier: skillTier,
    }),
  });
}

export interface CompletedMatch {
  id: number;
  court: string;
  team1: number[];
  team2: number[];
  team1_names: string[];
  team2_names: string[];
  scores: GameScore[];
  result: string;
  started_at: string;
  ended_at: string;
}

export async function getCompletedMatches(): Promise<APIResponse<CompletedMatch[]>> {
  return request<CompletedMatch[]>('/admin/matches/completed');
}

export async function getUserProfileById(userId: number): Promise<APIResponse<UserProfile>> {
  return request<UserProfile>(`/users/profile?id=${userId}`);
}

export async function getUserMatchHistory(userId: number): Promise<APIResponse<MatchHistory[]>> {
  return request<MatchHistory[]>(`/users/matches?id=${userId}`);
}

// Health check
export async function healthCheck(): Promise<boolean> {
  try {
    const response = await fetch(`${API_URL}/health`);
    const data = await response.json();
    return data.success === true;
  } catch {
    return false;
  }
}

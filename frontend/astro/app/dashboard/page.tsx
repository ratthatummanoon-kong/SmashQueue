"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import {
  getProfile,
  getQueueStatus,
  getMatchHistory,
  joinQueue,
  leaveQueue,
  getUser,
  getAccessToken,
  UserProfile,
  QueueInfo,
  MatchHistory,
} from "../lib/api";

export default function DashboardPage() {
  const [isLoading, setIsLoading] = useState(true);
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [queueInfo, setQueueInfo] = useState<QueueInfo | null>(null);
  const [matches, setMatches] = useState<MatchHistory[]>([]);
  const [isJoiningQueue, setIsJoiningQueue] = useState(false);
  const [error, setError] = useState("");
  const [showAllMatches, setShowAllMatches] = useState(false);

  useEffect(() => {
    const token = getAccessToken();
    if (!token) {
      window.location.href = "/login";
      return;
    }
    loadData();
  }, []);

  const loadData = async () => {
    setIsLoading(true);
    try {
      const [profileRes, queueRes, matchRes] = await Promise.all([
        getProfile(),
        getQueueStatus(),
        getMatchHistory(),
      ]);

      if (profileRes.success && profileRes.data) {
        setProfile(profileRes.data);
      }
      if (queueRes.success && queueRes.data) {
        setQueueInfo(queueRes.data);
      }
      if (matchRes.success && matchRes.data) {
        setMatches(matchRes.data); // Show all matches now
      }
    } catch (err) {
      console.error("Failed to load data:", err);
      setError("Failed to load dashboard data");
    } finally {
      setIsLoading(false);
    }
  };

  const handleJoinQueue = async () => {
    setIsJoiningQueue(true);
    try {
      const response = await joinQueue();
      if (response.success && response.data) {
        setQueueInfo(response.data.info);
      } else {
        setError(response.error?.message || "Failed to join queue");
      }
    } catch {
      setError("Network error");
    } finally {
      setIsJoiningQueue(false);
    }
  };

  const handleLeaveQueue = async () => {
    setIsJoiningQueue(true);
    try {
      const response = await leaveQueue();
      if (response.success && response.data) {
        setQueueInfo(response.data.info);
      }
    } catch {
      setError("Network error");
    } finally {
      setIsJoiningQueue(false);
    }
  };

  const user = getUser();
  const stats = profile?.stats || {
    win_rate: 0,
    total_matches: 0,
    current_streak: 0,
    skill_level: "Beginner",
  };

  if (isLoading) {
    return (
      <div className="min-h-screen pt-24 pb-12 px-4 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin w-12 h-12 border-4 border-[var(--primary)] border-t-transparent rounded-full mx-auto mb-4"></div>
          <p className="text-[var(--muted)]">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen pt-24 pb-12 px-4">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8 animate-fade-in">
          <h1 className="text-3xl font-bold mb-2">Dashboard</h1>
          <p className="text-[var(--muted)]">
            Welcome back, {user?.name || user?.username || "Player"}!
          </p>
        </div>

        {error && (
          <div className="mb-6 p-3 rounded-lg bg-[var(--error)]/10 border border-[var(--error)]/20 text-[var(--error)] text-sm">
            {error}
            <button onClick={() => setError("")} className="ml-2 underline">Dismiss</button>
          </div>
        )}

        {/* Stats Grid */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          <div className="card">
            <span className="text-[var(--muted)] text-sm">Win Rate</span>
            <p className="text-3xl font-bold text-[var(--success)] mt-2">
              {stats.win_rate?.toFixed(0) || 0}%
            </p>
          </div>
          <div className="card">
            <span className="text-[var(--muted)] text-sm">Matches</span>
            <p className="text-3xl font-bold mt-2">{stats.total_matches || 0}</p>
          </div>
          <div className="card">
            <span className="text-[var(--muted)] text-sm">Streak</span>
            <p className="text-3xl font-bold text-[var(--accent)] mt-2">
              {Math.abs(stats.current_streak || 0)}
            </p>
          </div>
          <div className="card">
            <span className="text-[var(--muted)] text-sm">Skill Level</span>
            <p className="text-xl font-bold mt-2">{stats.skill_level || "Beginner"}</p>
          </div>
        </div>

        <div className="grid lg:grid-cols-3 gap-6">
          {/* Queue Status */}
          <div className="lg:col-span-1">
            <div className="card">
              <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
                <span className="w-2 h-2 rounded-full bg-[var(--success)] animate-pulse"></span>
                Queue Status
              </h2>
              <div className="space-y-4">
                <div className="flex justify-between py-3 border-b border-[var(--border)]">
                  <span className="text-[var(--muted)]">Your Position</span>
                  <span className="text-2xl font-bold gradient-text">
                    {queueInfo?.your_position ? `#${queueInfo.your_position}` : "—"}
                  </span>
                </div>
                <div className="flex justify-between py-3 border-b border-[var(--border)]">
                  <span className="text-[var(--muted)]">Estimated Wait</span>
                  <span className="font-medium">{queueInfo?.estimated_wait || "—"}</span>
                </div>
                <div className="flex justify-between py-3">
                  <span className="text-[var(--muted)]">In Queue</span>
                  <span className="font-medium">{queueInfo?.total_in_queue || 0}</span>
                </div>
              </div>
              {queueInfo?.your_position ? (
                <button onClick={handleLeaveQueue} disabled={isJoiningQueue} className="btn-secondary w-full mt-4 disabled:opacity-50">
                  {isJoiningQueue ? "Leaving..." : "Leave Queue"}
                </button>
              ) : (
                <button onClick={handleJoinQueue} disabled={isJoiningQueue} className="btn-primary w-full mt-4 disabled:opacity-50">
                  {isJoiningQueue ? "Joining..." : "Join Queue"}
                </button>
              )}
            </div>
          </div>

          {/* Match History */}
          <div className="lg:col-span-2">
            <div className="card">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold">📜 Match History</h2>
                {matches.length > 5 && (
                  <button 
                    onClick={() => setShowAllMatches(!showAllMatches)}
                    className="text-sm text-[var(--primary)] hover:underline"
                  >
                    {showAllMatches ? "Show Less" : `View All (${matches.length})`}
                  </button>
                )}
              </div>
              <div className="space-y-3 max-h-[600px] overflow-y-auto">
                {matches.length === 0 ? (
                  <p className="text-center py-8 text-[var(--muted)]">No matches yet</p>
                ) : (
                  (showAllMatches ? matches : matches.slice(0, 5)).map((m) => (
                    <div key={m.match.id} className="flex items-start gap-4 p-3 rounded-lg bg-[var(--surface)] hover:bg-[var(--surface)]/80 transition-colors">
                      <div className={`w-12 h-12 rounded-lg flex items-center justify-center text-white font-bold flex-shrink-0 ${m.won ? "bg-[var(--success)]" : "bg-[var(--error)]"}`}>
                        {m.won ? "W" : "L"}
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex justify-between items-start mb-1">
                          <p className="font-medium">{m.match.court}</p>
                          <span className={`text-xs px-2 py-1 rounded ${
                            m.match.result === "team1" ? "bg-blue-500/20 text-blue-400" :
                            m.match.result === "team2" ? "bg-orange-500/20 text-orange-400" :
                            "bg-gray-500/20 text-gray-400"
                          }`}>
                            {m.match.result === "team1" ? "Team 1 Won" : 
                             m.match.result === "team2" ? "Team 2 Won" : "Draw"}
                          </span>
                        </div>
                        <p className="text-sm text-[var(--muted)]">
                          {new Date(m.match.started_at).toLocaleDateString()} • {new Date(m.match.started_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                        </p>
                        {m.match.scores && m.match.scores.length > 0 && (
                          <p className="text-xs text-[var(--muted)] mt-1">
                            Scores: {m.match.scores.map((s: any) => `${s.team1_score}-${s.team2_score}`).join(", ")}
                          </p>
                        )}
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

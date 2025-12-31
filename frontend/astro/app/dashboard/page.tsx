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
  getAllUsers,
  updatePlayerAdmin,
  getUserProfileById,
  getUserMatchHistory,
  isAdmin,
  UserProfile,
  QueueInfo,
  MatchHistory,
  UserListItem,
} from "../lib/api";

export default function DashboardPage() {
  const [isLoading, setIsLoading] = useState(true);
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [queueInfo, setQueueInfo] = useState<QueueInfo | null>(null);
  const [matches, setMatches] = useState<MatchHistory[]>([]);
  const [isJoiningQueue, setIsJoiningQueue] = useState(false);
  const [error, setError] = useState("");
  const [showAllMatches, setShowAllMatches] = useState(false);
  const [allUsers, setAllUsers] = useState<UserListItem[]>([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [editingPlayer, setEditingPlayer] = useState<UserListItem | null>(null);
  const [editForm, setEditForm] = useState({
    handPreference: "right",
    skillTier: "N",
  });
  const [viewingUserStats, setViewingUserStats] = useState<UserListItem | null>(null);
  const [userStatsData, setUserStatsData] = useState<any>(null);
  const [userMatchHistory, setUserMatchHistory] = useState<any[]>([]);
  const [success, setSuccess] = useState("");
  const [playersPerPage, setPlayersPerPage] = useState(20);
  const [customPlayersPerPage, setCustomPlayersPerPage] = useState("");

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
    const user = getUser();
    const adminView = user && isAdmin();
    
    console.log("Loading dashboard, user:", user, "adminView:", adminView);
    
    try {
      if (adminView) {
        console.log("Fetching all users for admin view...");
        const usersRes = await getAllUsers();
        console.log("Users response:", usersRes);
        if (usersRes.success && usersRes.data) {
          setAllUsers(usersRes.data);
        } else {
          setError(usersRes.error?.message || "Failed to load users");
        }
      } else {
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
          setMatches(matchRes.data);
        }
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

  const openEditPlayer = (player: UserListItem) => {
    setEditingPlayer(player);
    setEditForm({
      handPreference: player.hand_preference || "right",
      skillTier: player.skill_tier || "N",
    });
  };

  const handleUpdatePlayer = async () => {
    if (!editingPlayer) return;

    try {
      const response = await updatePlayerAdmin(
        editingPlayer.id,
        editForm.handPreference,
        editForm.skillTier
      );
      
      if (response.success) {
        setSuccess(`Updated ${editingPlayer.name}'s settings`);
        setEditingPlayer(null);
        loadData();
        setTimeout(() => setSuccess(""), 3000);
      } else {
        setError(response.error?.message || "Failed to update player");
      }
    } catch {
      setError("Network error");
    }
  };

  const viewUserStats = async (player: UserListItem) => {
    setViewingUserStats(player);
    try {
      const [profileRes, matchesRes] = await Promise.all([
        getUserProfileById(player.id),
        getUserMatchHistory(player.id),
      ]);
      
      if (profileRes.success && profileRes.data) {
        setUserStatsData(profileRes.data);
      }
      if (matchesRes.success && matchesRes.data) {
        setUserMatchHistory(matchesRes.data);
      }
    } catch (err) {
      console.error("Failed to load user stats:", err);
    }
  };

  const user = getUser();
  const adminView = user && isAdmin();
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
            {adminView ? "Admin View - All Players Directory" : `Welcome back, ${user?.name || user?.username || "Player"}!`}
          </p>
        </div>

        {error && (
          <div className="mb-6 p-3 rounded-lg bg-[var(--error)]/10 border border-[var(--error)]/20 text-[var(--error)] text-sm">
            {error}
            <button onClick={() => setError("")} className="ml-2 underline">Dismiss</button>
          </div>
        )}

        {success && (
          <div className="mb-6 p-3 rounded-lg bg-[var(--success)]/10 border border-[var(--success)]/20 text-[var(--success)] text-sm">
            {success}
          </div>
        )}

        {adminView ? (
          /* Admin View - Player Directory */
          <div className="card">
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
              <h2 className="text-2xl font-semibold">
                👥 All Players ({Math.min(playersPerPage, allUsers.filter(u => 
                  searchTerm === "" || 
                  u.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                  u.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
                  u.phone?.includes(searchTerm)
                ).length)}/{allUsers.filter(u => 
                  searchTerm === "" || 
                  u.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                  u.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
                  u.phone?.includes(searchTerm)
                ).length})
              </h2>
              <input
                type="text"
                placeholder="Search by name, username, or phone..."
                className="input w-full sm:w-64"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </div>

            {/* Pagination Controls */}
            <div className="flex flex-wrap items-center gap-2 mb-4 pb-4 border-b border-[var(--border)]">
              <span className="text-sm text-[var(--muted)]">Show:</span>
              {[10, 20, 50, 100].map(num => (
                <button
                  key={num}
                  onClick={() => setPlayersPerPage(num)}
                  className={`text-xs px-3 py-1 rounded transition-colors ${
                    playersPerPage === num 
                      ? 'bg-[var(--primary)] text-white' 
                      : 'bg-[var(--surface)] hover:bg-[var(--border)]'
                  }`}
                >
                  {num}
                </button>
              ))}
              <input
                type="number"
                placeholder="Custom"
                className="input text-xs w-20 py-1"
                value={customPlayersPerPage}
                onChange={(e) => setCustomPlayersPerPage(e.target.value)}
                onKeyPress={(e) => {
                  if (e.key === 'Enter' && customPlayersPerPage) {
                    const num = parseInt(customPlayersPerPage);
                    if (num > 0) setPlayersPerPage(num);
                  }
                }}
              />
              <button
                onClick={() => {
                  const num = parseInt(customPlayersPerPage);
                  if (num > 0) setPlayersPerPage(num);
                }}
                className="text-xs px-3 py-1 bg-[var(--surface)] hover:bg-[var(--border)] rounded"
                disabled={!customPlayersPerPage}
              >
                Apply
              </button>
            </div>
            
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-[var(--border)]">
                    <th className="text-left py-3 px-2">Name</th>
                    <th className="text-left py-3 px-2 hidden sm:table-cell">Phone</th>
                    <th className="text-left py-3 px-2">Hand</th>
                    <th className="text-left py-3 px-2">Tier</th>
                    <th className="text-left py-3 px-2">Matches</th>
                    <th className="text-left py-3 px-2">Win Rate</th>
                    <th className="text-left py-3 px-2">Wins</th>
                    <th className="text-left py-3 px-2">Losses</th>
                    <th className="text-right py-3 px-2">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {allUsers
                    .filter(u => 
                      searchTerm === "" || 
                      u.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                      u.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
                      u.phone?.includes(searchTerm)
                    )
                    .slice(0, playersPerPage)
                    .map(player => (
                      <tr key={player.id} className="border-b border-[var(--border)]/50 hover:bg-[var(--surface)]">
                        <td className="py-3 px-2">
                          <div>
                            <p className="font-medium">{player.name}</p>
                            <p className="text-xs text-[var(--muted)]">@{player.username}</p>
                          </div>
                        </td>
                        <td className="py-3 px-2 hidden sm:table-cell text-[var(--muted)]">{player.phone || "—"}</td>
                        <td className="py-3 px-2">
                          <span className="text-sm">
                            {player.hand_preference === "left" ? "👈 Left" : "👉 Right"}
                          </span>
                        </td>
                        <td className="py-3 px-2">
                          <span className={`text-xs px-2 py-1 rounded font-bold ${
                            ["A", "B", "C"].includes(player.skill_tier || "") ? 'bg-purple-500/20 text-purple-400' :
                            ["P+", "P", "P-"].includes(player.skill_tier || "") ? 'bg-blue-500/20 text-blue-400' :
                            ["N", "S", "S-"].includes(player.skill_tier || "") ? 'bg-green-500/20 text-green-400' :
                            'bg-gray-500/20 text-gray-400'
                          }`}>
                            {player.skill_tier || "N"}
                          </span>
                        </td>
                        <td className="py-3 px-2 font-medium">{player.total_matches || 0}</td>
                        <td className="py-3 px-2">
                          <span className={player.win_rate >= 50 ? 'text-[var(--success)] font-medium' : 'text-[var(--muted)]'}>
                            {player.win_rate?.toFixed(0) || 0}%
                          </span>
                        </td>
                        <td className="py-3 px-2 text-green-500 font-medium">{player.wins || 0}</td>
                        <td className="py-3 px-2 text-red-500 font-medium">
                          {(player.total_matches || 0) - (player.wins || 0)}
                        </td>
                        <td className="py-3 px-2 text-right">
                          <div className="flex gap-1 justify-end">
                            <button 
                              onClick={() => viewUserStats(player)}
                              className="text-xs px-3 py-1 bg-green-500/20 text-green-400 rounded hover:bg-green-500/30"
                              title="View player stats"
                            >
                              📊
                            </button>
                            <button 
                              onClick={() => openEditPlayer(player)}
                              className="text-xs px-3 py-1 bg-purple-500/20 text-purple-400 rounded hover:bg-purple-500/30"
                              title="Edit player settings"
                            >
                              ✏️
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                </tbody>
              </table>
            </div>
          </div>
        ) : (
          /* Regular Player View */
          <>
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
        </>
        )}

        {/* Edit Player Modal */}
        {editingPlayer && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
            <div className="bg-[var(--card)] rounded-xl p-6 max-w-md w-full">
              <h3 className="text-xl font-bold mb-4">Edit Player: {editingPlayer.name}</h3>
              
              <div className="space-y-4">
                <div>
                  <label className="label">Hand Preference</label>
                  <div className="grid grid-cols-2 gap-2">
                    <button
                      onClick={() => setEditForm({ ...editForm, handPreference: "right" })}
                      className={`p-3 rounded-lg border-2 transition-colors ${
                        editForm.handPreference === "right"
                          ? "border-[var(--primary)] bg-[var(--primary)]/10"
                          : "border-[var(--border)] hover:border-[var(--border-hover)]"
                      }`}
                    >
                      👉 Right
                    </button>
                    <button
                      onClick={() => setEditForm({ ...editForm, handPreference: "left" })}
                      className={`p-3 rounded-lg border-2 transition-colors ${
                        editForm.handPreference === "left"
                          ? "border-[var(--primary)] bg-[var(--primary)]/10"
                          : "border-[var(--border)] hover:border-[var(--border-hover)]"
                      }`}
                    >
                      👈 Left
                    </button>
                  </div>
                </div>

                <div>
                  <label className="label">Skill Tier</label>
                  <select
                    className="input"
                    value={editForm.skillTier}
                    onChange={(e) => setEditForm({ ...editForm, skillTier: e.target.value })}
                  >
                    <option value="BG">BG - Beginner</option>
                    <option value="S-">S- - Semi Novice</option>
                    <option value="S">S - Semi</option>
                    <option value="N">N - Novice</option>
                    <option value="P-">P- - Pre-Professional</option>
                    <option value="P">P - Professional</option>
                    <option value="P+">P+ - Pro Plus</option>
                    <option value="C">C - Champion</option>
                    <option value="B">B - Bronze</option>
                    <option value="A">A - Ace</option>
                  </select>
                </div>
              </div>

              <div className="flex gap-3 mt-6">
                <button
                  onClick={() => setEditingPlayer(null)}
                  className="btn-secondary flex-1"
                >
                  Cancel
                </button>
                <button
                  onClick={handleUpdatePlayer}
                  className="btn-primary flex-1"
                >
                  💾 Save Changes
                </button>
              </div>
            </div>
          </div>
        )}

        {/* View User Stats Modal */}
        {viewingUserStats && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
            <div className="bg-[var(--card)] rounded-xl p-6 max-w-2xl w-full max-h-[90vh] overflow-y-auto">
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-xl font-bold">📊 Player Statistics: {viewingUserStats.name}</h3>
                <button onClick={() => setViewingUserStats(null)} className="text-2xl hover:text-red-500">×</button>
              </div>

              {userStatsData && (
                <div className="grid grid-cols-2 gap-4 mb-6">
                  <div className="card">
                    <p className="text-sm text-[var(--muted)]">Skill Level</p>
                    <p className="text-2xl font-bold">{userStatsData.user?.skill_tier || userStatsData.stats?.skill_level}</p>
                  </div>
                  <div className="card">
                    <p className="text-sm text-[var(--muted)]">Rating</p>
                    <p className="text-2xl font-bold">{userStatsData.user?.rating?.toFixed(1) || "N/A"}</p>
                  </div>
                  <div className="card">
                    <p className="text-sm text-[var(--muted)]">Matches Played</p>
                    <p className="text-2xl font-bold">{userStatsData.stats?.total_matches || 0}</p>
                  </div>
                  <div className="card">
                    <p className="text-sm text-[var(--muted)]">Win Rate</p>
                    <p className="text-2xl font-bold">
                      {userStatsData.stats?.total_matches 
                        ? ((userStatsData.stats.wins / userStatsData.stats.total_matches) * 100).toFixed(1) 
                        : "0"}%
                    </p>
                  </div>
                  <div className="card">
                    <p className="text-sm text-[var(--muted)]">Wins</p>
                    <p className="text-2xl font-bold text-green-500">{userStatsData.stats?.wins || 0}</p>
                  </div>
                  <div className="card">
                    <p className="text-sm text-[var(--muted)]">Losses</p>
                    <p className="text-2xl font-bold text-red-500">{userStatsData.stats?.losses || 0}</p>
                  </div>
                </div>
              )}

              <h4 className="font-semibold mb-3">Match History</h4>
              <div className="space-y-2 max-h-96 overflow-y-auto">
                {userMatchHistory.length > 0 ? (
                  userMatchHistory.map((match: any) => (
                    <div key={match.id} className="p-3 bg-[var(--surface)] rounded">
                      <div className="flex justify-between items-center mb-2">
                        <span className="text-sm font-medium">{match.court}</span>
                        <span className={`text-xs px-2 py-1 rounded ${
                          match.won ? "bg-green-500/20 text-green-400" : "bg-red-500/20 text-red-400"
                        }`}>
                          {match.won ? "Won" : "Lost"}
                        </span>
                      </div>
                      <div className="text-xs text-[var(--muted)]">
                        <p>Team 1: {match.team1_names?.join(", ") || "N/A"}</p>
                        <p>Team 2: {match.team2_names?.join(", ") || "N/A"}</p>
                        {match.scores && match.scores.length > 0 && (
                          <p className="mt-1">
                            Scores: {match.scores.map((s: any) => `${s.team1_score}-${s.team2_score}`).join(", ")}
                          </p>
                        )}
                        <p className="mt-1">{new Date(match.started_at).toLocaleDateString()}</p>
                      </div>
                    </div>
                  ))
                ) : (
                  <p className="text-center text-[var(--muted)] py-4">No match history</p>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

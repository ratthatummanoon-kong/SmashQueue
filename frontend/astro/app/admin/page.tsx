"use client";

import { useState, useEffect } from "react";
import {
  getAccessToken,
  getQueueStatus,
  createMatch,
  getActiveMatches,
  callNextPlayers,
  getAllUsers,
  updatePlayerAdmin,
  getCompletedMatches,
  isAdmin,
  isOrganizer,
  QueueInfo,
  Match,
  UserListItem,
  CompletedMatch,
} from "../lib/api";

export default function AdminPage() {
  const [isLoading, setIsLoading] = useState(true);
  const [queueInfo, setQueueInfo] = useState<QueueInfo | null>(null);
  const [activeMatches, setActiveMatches] = useState<Match[]>([]);
  const [completedMatches, setCompletedMatches] = useState<CompletedMatch[]>([]);
  const [allUsers, setAllUsers] = useState<UserListItem[]>([]);
  const [selectedTeam1, setSelectedTeam1] = useState<UserListItem[]>([]);
  const [selectedTeam2, setSelectedTeam2] = useState<UserListItem[]>([]);
  const [court, setCourt] = useState("Court 1");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [isCreating, setIsCreating] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [editingPlayer, setEditingPlayer] = useState<UserListItem | null>(null);
  const [editForm, setEditForm] = useState({
    handPreference: "right",
    skillTier: "N",
  });

  const skillTiers = ["BG", "S-", "S", "N", "P-", "P", "P+", "C", "B", "A"];
  const skillTierNames: Record<string, string> = {
    BG: "Beginner",
    "S-": "Sub-Standard",
    S: "Standard",
    N: "Normal",
    "P-": "Pro Minus",
    P: "Pro",
    "P+": "Pro Plus",
    C: "Champion",
    B: "Best",
    A: "Ace",
  };

  useEffect(() => {
    const token = getAccessToken();
    if (!token) {
      window.location.href = "/login";
      return;
    }
    if (!isAdmin() && !isOrganizer()) {
      window.location.href = "/dashboard";
      return;
    }
    loadData();
  }, []);

  const loadData = async () => {
    setIsLoading(true);
    setError("");
    try {
      const [queueRes, matchRes, usersRes, completedRes] = await Promise.all([
        getQueueStatus(),
        getActiveMatches(),
        getAllUsers(),
        getCompletedMatches(),
      ]);

      console.log("Queue response:", queueRes);
      console.log("Match response:", matchRes);
      console.log("Users response:", usersRes);
      console.log("Completed matches:", completedRes);

      if (queueRes.success && queueRes.data) {
        setQueueInfo(queueRes.data);
      } else {
        console.error("Queue fetch failed:", queueRes.error);
      }
      
      if (matchRes.success && matchRes.data) {
        setActiveMatches(matchRes.data);
      } else {
        console.error("Match fetch failed:", matchRes.error);
      }
      
      if (usersRes.success && usersRes.data) {
        console.log("Loaded users:", usersRes.data.length);
        setAllUsers(usersRes.data);
      } else {
        console.error("Users fetch failed:", usersRes.error);
        setError(usersRes.error?.message || "Failed to load users");
      }

      if (completedRes.success && completedRes.data) {
        setCompletedMatches(completedRes.data);
      } else {
        console.error("Completed matches fetch failed:", completedRes.error);
      }
    } catch (err) {
      console.error("Failed to load:", err);
      setError("Failed to load data: " + (err instanceof Error ? err.message : "Unknown error"));
    } finally {
      setIsLoading(false);
    }
  };

  const handleCallNext = async () => {
    try {
      const response = await callNextPlayers();
      if (response.success && response.data) {
        setSuccess(`Called ${response.data.count} players from queue`);
        loadData();
        setTimeout(() => setSuccess(""), 3000);
      } else {
        setError(response.error?.message || "Failed to call players");
      }
    } catch {
      setError("Network error");
    }
  };

  const addToTeam1 = (user: UserListItem) => {
    if (selectedTeam1.length < 2 && !isInAnyTeam(user.id)) {
      setSelectedTeam1([...selectedTeam1, user]);
    }
  };

  const addToTeam2 = (user: UserListItem) => {
    if (selectedTeam2.length < 2 && !isInAnyTeam(user.id)) {
      setSelectedTeam2([...selectedTeam2, user]);
    }
  };

  const removeFromTeam1 = (userId: number) => {
    setSelectedTeam1(selectedTeam1.filter(u => u.id !== userId));
  };

  const removeFromTeam2 = (userId: number) => {
    setSelectedTeam2(selectedTeam2.filter(u => u.id !== userId));
  };

  const isInAnyTeam = (userId: number) => {
    return selectedTeam1.some(u => u.id === userId) || selectedTeam2.some(u => u.id === userId);
  };

  const moveToTeam1 = (user: UserListItem) => {
    removeFromTeam2(user.id);
    if (selectedTeam1.length < 2) {
      setSelectedTeam1([...selectedTeam1, user]);
    }
  };

  const moveToTeam2 = (user: UserListItem) => {
    removeFromTeam1(user.id);
    if (selectedTeam2.length < 2) {
      setSelectedTeam2([...selectedTeam2, user]);
    }
  };

  const canCreateMatch = selectedTeam1.length >= 1 && selectedTeam2.length >= 1;

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

  const handleCreateMatch = async () => {
    if (!canCreateMatch) {
      setError("Please select at least 1 player per team");
      return;
    }

    setIsCreating(true);
    try {
      const response = await createMatch(
        court,
        selectedTeam1.map(u => u.id),
        selectedTeam2.map(u => u.id)
      );
      if (response.success && response.data) {
        setSuccess(`Match created on ${court}!`);
        setSelectedTeam1([]);
        setSelectedTeam2([]);
        loadData();
        setTimeout(() => setSuccess(""), 3000);
      } else {
        setError(response.error?.message || "Failed to create match");
      }
    } catch {
      setError("Network error");
    } finally {
      setIsCreating(false);
    }
  };

  const resetTeams = () => {
    setSelectedTeam1([]);
    setSelectedTeam2([]);
  };

  // Filter users by search
  const filteredUsers = allUsers.filter(u => 
    !isInAnyTeam(u.id) && u.role === 'player' &&
    (u.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
     u.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
     u.phone.includes(searchTerm))
  );

  if (isLoading) {
    return (
      <div className="min-h-screen pt-24 pb-12 px-4 flex items-center justify-center">
        <div className="animate-spin w-12 h-12 border-4 border-[var(--primary)] border-t-transparent rounded-full"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen pt-24 pb-12 px-4">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold mb-2">🏸 Admin Panel</h1>
          <p className="text-[var(--muted)]">Manage queue, create matches, and view player data</p>
        </div>

        {error && (
          <div className="mb-6 p-3 rounded-lg bg-[var(--error)]/10 text-[var(--error)] text-sm flex justify-between">
            {error}
            <button onClick={() => setError("")}>×</button>
          </div>
        )}
        {success && (
          <div className="mb-6 p-3 rounded-lg bg-[var(--success)]/10 text-[var(--success)] text-sm">
            ✓ {success}
          </div>
        )}

        <div className="grid lg:grid-cols-3 gap-6">
          {/* Queue Stats */}
          <div className="card">
            <h2 className="text-lg font-semibold mb-4">📋 Queue Status</h2>
            <div className="space-y-3 mb-4">
              <div className="flex justify-between py-2 border-b border-[var(--border)]">
                <span className="text-[var(--muted)]">In Queue</span>
                <span className="font-bold text-2xl gradient-text">{queueInfo?.total_in_queue || 0}</span>
              </div>
              <div className="flex justify-between py-2">
                <span className="text-[var(--muted)]">Total Players</span>
                <span className="font-medium">{allUsers.filter(u => u.role === 'player').length}</span>
              </div>
            </div>
            <button onClick={handleCallNext} className="btn-primary w-full">
              📢 Call Next 4 Players
            </button>
          </div>

          {/* Match Creation */}
          <div className="lg:col-span-2 card">
            <h2 className="text-lg font-semibold mb-4">⚔️ Create Match</h2>

            {/* Court Selection */}
            <div className="mb-4">
              <label className="label">Court</label>
              <select 
                className="input"
                value={court}
                onChange={(e) => setCourt(e.target.value)}
              >
                <option>Court 1</option>
                <option>Court 2</option>
                <option>Court 3</option>
                <option>Court 4</option>
              </select>
            </div>

            {/* Team Selection Area */}
            <div className="grid md:grid-cols-2 gap-4 mb-6">
              {/* Team 1 */}
              <div className="p-4 rounded-lg border-2 border-blue-500/30 bg-blue-500/5">
                <h3 className="font-semibold text-blue-400 mb-3">🔵 Team 1</h3>
                <div className="space-y-2 min-h-[80px]">
                  {selectedTeam1.map(user => (
                    <div key={user.id} className="flex justify-between items-center p-2 bg-blue-500/10 rounded">
                      <div>
                        <span className="font-medium">{user.name}</span>
                        <span className="text-xs text-[var(--muted)] ml-2">{user.skill_level}</span>
                      </div>
                      <div className="flex gap-1">
                        <button onClick={() => moveToTeam2(user)} className="text-xs px-2 py-1 bg-orange-500/20 rounded hover:bg-orange-500/30">→</button>
                        <button onClick={() => removeFromTeam1(user.id)} className="text-xs px-2 py-1 bg-red-500/20 rounded hover:bg-red-500/30">×</button>
                      </div>
                    </div>
                  ))}
                  {selectedTeam1.length === 0 && (
                    <p className="text-[var(--muted)] text-sm text-center py-4">Select players below</p>
                  )}
                </div>
              </div>

              {/* Team 2 */}
              <div className="p-4 rounded-lg border-2 border-orange-500/30 bg-orange-500/5">
                <h3 className="font-semibold text-orange-400 mb-3">🟠 Team 2</h3>
                <div className="space-y-2 min-h-[80px]">
                  {selectedTeam2.map(user => (
                    <div key={user.id} className="flex justify-between items-center p-2 bg-orange-500/10 rounded">
                      <div>
                        <span className="font-medium">{user.name}</span>
                        <span className="text-xs text-[var(--muted)] ml-2">{user.skill_level}</span>
                      </div>
                      <div className="flex gap-1">
                        <button onClick={() => moveToTeam1(user)} className="text-xs px-2 py-1 bg-blue-500/20 rounded hover:bg-blue-500/30">←</button>
                        <button onClick={() => removeFromTeam2(user.id)} className="text-xs px-2 py-1 bg-red-500/20 rounded hover:bg-red-500/30">×</button>
                      </div>
                    </div>
                  ))}
                  {selectedTeam2.length === 0 && (
                    <p className="text-[var(--muted)] text-sm text-center py-4">Select players below</p>
                  )}
                </div>
              </div>
            </div>

            {/* Match Preview */}
            {(selectedTeam1.length > 0 || selectedTeam2.length > 0) && (
              <div className="mb-4 p-4 rounded-lg bg-gradient-to-r from-blue-500/10 via-[var(--surface)] to-orange-500/10 text-center border border-[var(--border)]">
                <p className="text-lg font-bold">
                  <span className="text-blue-400">{selectedTeam1.map(u => u.name).join(" & ") || "?"}</span>
                  <span className="mx-3 text-[var(--muted)]">⚔️</span>
                  <span className="text-orange-400">{selectedTeam2.map(u => u.name).join(" & ") || "?"}</span>
                </p>
                <p className="text-sm text-[var(--muted)] mt-1">{court}</p>
              </div>
            )}

            {/* Actions */}
            <div className="flex gap-3">
              <button onClick={resetTeams} className="btn-secondary flex-1">
                Reset
              </button>
              <button 
                onClick={handleCreateMatch} 
                disabled={!canCreateMatch || isCreating}
                className="btn-primary flex-1 disabled:opacity-50"
              >
                {isCreating ? "Creating..." : "🏸 Start Match"}
              </button>
            </div>
          </div>
        </div>

        {/* Player Directory */}
        <div className="card mt-6">
          <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-4">
            <h2 className="text-lg font-semibold">👥 Player Directory</h2>
            <input
              type="text"
              placeholder="Search by name, username, or phone..."
              className="input w-full sm:w-64"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
          
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-[var(--border)]">
                  <th className="text-left py-3 px-2">Name</th>
                  <th className="text-left py-3 px-2 hidden sm:table-cell">Phone</th>
                  <th className="text-left py-3 px-2">Hand</th>
                  <th className="text-left py-3 px-2">Tier</th>
                  <th className="text-left py-3 px-2">Win Rate</th>
                  <th className="text-right py-3 px-2">Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredUsers.slice(0, 20).map(user => (
                  <tr key={user.id} className="border-b border-[var(--border)]/50 hover:bg-[var(--surface)]">
                    <td className="py-3 px-2">
                      <div>
                        <p className="font-medium">{user.name}</p>
                        <p className="text-xs text-[var(--muted)]">@{user.username}</p>
                      </div>
                    </td>
                    <td className="py-3 px-2 hidden sm:table-cell text-[var(--muted)]">{user.phone || "—"}</td>
                    <td className="py-3 px-2">
                      <span className="text-sm">
                        {user.hand_preference === "left" ? "👈 Left" : "👉 Right"}
                      </span>
                    </td>
                    <td className="py-3 px-2">
                      <span className={`text-xs px-2 py-1 rounded font-bold ${
                        ["A", "B", "C"].includes(user.skill_tier || "") ? 'bg-purple-500/20 text-purple-400' :
                        ["P+", "P", "P-"].includes(user.skill_tier || "") ? 'bg-blue-500/20 text-blue-400' :
                        ["N", "S", "S-"].includes(user.skill_tier || "") ? 'bg-green-500/20 text-green-400' :
                        'bg-gray-500/20 text-gray-400'
                      }`}>
                        {user.skill_tier || "N"}
                      </span>
                    </td>
                    <td className="py-3 px-2">
                      <span className={user.win_rate >= 50 ? 'text-[var(--success)]' : 'text-[var(--muted)]'}>
                        {user.win_rate?.toFixed(0) || 0}%
                      </span>
                    </td>
                    <td className="py-3 px-2 text-right">
                      <div className="flex gap-1 justify-end">
                        <button 
                          onClick={() => openEditPlayer(user)}
                          className="text-xs px-3 py-1 bg-purple-500/20 text-purple-400 rounded hover:bg-purple-500/30"
                          title="Edit player settings"
                        >
                          ✏️
                        </button>
                        <button 
                          onClick={() => addToTeam1(user)}
                          disabled={selectedTeam1.length >= 2}
                          className="text-xs px-3 py-1 bg-blue-500/20 text-blue-400 rounded hover:bg-blue-500/30 disabled:opacity-50"
                        >
                          🔵
                        </button>
                        <button 
                          onClick={() => addToTeam2(user)}
                          disabled={selectedTeam2.length >= 2}
                          className="text-xs px-3 py-1 bg-orange-500/20 text-orange-400 rounded hover:bg-orange-500/30 disabled:opacity-50"
                        >
                          🟠
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {filteredUsers.length === 0 && allUsers.length === 0 && (
              <div className="text-center py-8">
                <p className="text-[var(--muted)] mb-2">No players found in database</p>
                <p className="text-sm text-[var(--muted)]">Players will appear here after they register</p>
              </div>
            )}
            {filteredUsers.length === 0 && allUsers.length > 0 && (
              <p className="text-center py-8 text-[var(--muted)]">No players match your search</p>
            )}
            {filteredUsers.length > 20 && (
              <p className="text-center py-4 text-[var(--muted)] text-sm">Showing 20 of {filteredUsers.length} players</p>
            )}
          </div>
        </div>

        {/* Active Matches */}
        <div className="card mt-6">
          <h2 className="text-lg font-semibold mb-4">🎮 Active Matches ({activeMatches.length})</h2>
          {activeMatches.length === 0 ? (
            <p className="text-[var(--muted)] text-center py-8">No active matches</p>
          ) : (
            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
              {activeMatches.map(match => (
                <div key={match.id} className="p-4 rounded-lg bg-[var(--surface)] border border-[var(--border)]">
                  <div className="flex justify-between items-center mb-2">
                    <span className="font-medium">{match.court}</span>
                    <span className="text-xs px-2 py-1 rounded bg-[var(--accent)]/20 text-[var(--accent)]">
                      In Progress
                    </span>
                  </div>
                  <p className="text-sm text-[var(--muted)]">
                    Team 1: {match.team1.join(", ")}
                    <br />
                    Team 2: {match.team2.join(", ")}
                  </p>
                  <p className="text-xs text-[var(--muted)] mt-2">
                    Started {new Date(match.started_at).toLocaleTimeString()}
                  </p>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Match History */}
        <div className="card mt-6">
          <h2 className="text-lg font-semibold mb-4">📊 Match History ({completedMatches.length})</h2>
          {completedMatches.length === 0 ? (
            <p className="text-[var(--muted)] text-center py-8">No completed matches yet</p>
          ) : (
            <div className="space-y-3">
              {completedMatches.slice(0, 20).map(match => {
                const team1 = match.team1_names.join(" & ");
                const team2 = match.team2_names.join(" & ");
                const scoresStr = match.scores
                  .map(s => `${s.team1_score}-${s.team2_score}`)
                  .join(" ");
                
                return (
                  <div
                    key={match.id}
                    className="p-4 rounded-lg bg-[var(--surface)] border border-[var(--border)] hover:border-[var(--primary)]/30 transition-colors"
                  >
                    <div className="flex flex-wrap items-center justify-between gap-2 mb-2">
                      <div className="flex items-center gap-2">
                        <span className="text-xs px-2 py-1 rounded bg-[var(--muted)]/20 text-[var(--muted)]">
                          {match.court}
                        </span>
                        <span className="text-xs text-[var(--muted)]">
                          {new Date(match.ended_at).toLocaleString()}
                        </span>
                      </div>
                      <span className={`text-xs px-2 py-1 rounded font-medium ${
                        match.result === 'team1' ? 'bg-blue-500/20 text-blue-400' :
                        match.result === 'team2' ? 'bg-orange-500/20 text-orange-400' :
                        'bg-gray-500/20 text-gray-400'
                      }`}>
                        {match.result === 'team1' ? 'Team 1 Won' :
                         match.result === 'team2' ? 'Team 2 Won' : 'Draw'}
                      </span>
                    </div>
                    <div className="text-sm">
                      <span className={match.result === 'team1' ? 'font-bold text-blue-400' : ''}>
                        {team1}
                      </span>
                      <span className="mx-2 text-[var(--muted)]">vs</span>
                      <span className={match.result === 'team2' ? 'font-bold text-orange-400' : ''}>
                        {team2}
                      </span>
                    </div>
                    {scoresStr && (
                      <div className="mt-2 text-xs text-[var(--muted)]">
                        Scores: <span className="font-mono">{scoresStr}</span>
                      </div>
                    )}
                  </div>
                );
              })}
              {completedMatches.length > 20 && (
                <p className="text-center py-4 text-[var(--muted)] text-sm">
                  Showing 20 of {completedMatches.length} matches
                </p>
              )}
            </div>
          )}
        </div>

        {/* Edit Player Modal */}
        {editingPlayer && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
            <div className="bg-[var(--surface-elevated)] rounded-lg max-w-md w-full p-6 animate-fade-in">
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-xl font-bold">✏️ Edit Player Settings</h2>
                <button
                  onClick={() => setEditingPlayer(null)}
                  className="text-[var(--muted)] hover:text-[var(--foreground)] text-2xl leading-none"
                >
                  ×
                </button>
              </div>

              <div className="mb-4">
                <p className="text-lg font-medium mb-1">{editingPlayer.name}</p>
                <p className="text-sm text-[var(--muted)]">@{editingPlayer.username}</p>
              </div>

              <div className="space-y-4">
                <div>
                  <label className="label">Dominant Hand</label>
                  <div className="grid grid-cols-2 gap-3">
                    <button
                      onClick={() => setEditForm({ ...editForm, handPreference: "right" })}
                      className={`py-3 px-4 rounded-lg border-2 transition-all ${
                        editForm.handPreference === "right"
                          ? "border-[var(--primary)] bg-[var(--primary)]/10"
                          : "border-[var(--border)] hover:border-[var(--primary)]/50"
                      }`}
                    >
                      <span className="text-2xl mb-1 block">👉</span>
                      <span className="text-sm">Right</span>
                    </button>
                    <button
                      onClick={() => setEditForm({ ...editForm, handPreference: "left" })}
                      className={`py-3 px-4 rounded-lg border-2 transition-all ${
                        editForm.handPreference === "left"
                          ? "border-[var(--primary)] bg-[var(--primary)]/10"
                          : "border-[var(--border)] hover:border-[var(--primary)]/50"
                      }`}
                    >
                      <span className="text-2xl mb-1 block">👈</span>
                      <span className="text-sm">Left</span>
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
                    {skillTiers.map((tier) => (
                      <option key={tier} value={tier}>
                        {tier} - {skillTierNames[tier]}
                      </option>
                    ))}
                  </select>
                  <p className="text-xs text-[var(--muted)] mt-2">
                    Current: <span className="font-bold">{editingPlayer.skill_tier || "N"}</span>
                  </p>
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
      </div>
    </div>
  );
}

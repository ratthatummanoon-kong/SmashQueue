"use client";

import { useState, useEffect } from "react";
import {
  getProfile,
  updateProfile,
  getAccessToken,
  logout,
  UserProfile,
} from "../lib/api";

export default function ProfilePage() {
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [profileData, setProfileData] = useState<UserProfile | null>(null);
  const [editForm, setEditForm] = useState({ name: "", bio: "" });
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  useEffect(() => {
    const token = getAccessToken();
    if (!token) {
      window.location.href = "/login";
      return;
    }
    loadProfile();
  }, []);

  const loadProfile = async () => {
    setIsLoading(true);
    try {
      const response = await getProfile();
      if (response.success && response.data) {
        setProfileData(response.data);
        setEditForm({ name: response.data.user.name, bio: response.data.user.bio });
      }
    } catch {
      setError("Failed to load profile");
    } finally {
      setIsLoading(false);
    }
  };

  const handleSave = async () => {
    setIsSaving(true);
    try {
      const response = await updateProfile(editForm.name, editForm.bio);
      if (response.success && response.data) {
        setProfileData((prev) => prev ? { ...prev, user: response.data! } : null);
        setIsEditing(false);
        setSuccess("Profile updated!");
        setTimeout(() => setSuccess(""), 3000);
      }
    } catch {
      setError("Failed to update");
    } finally {
      setIsSaving(false);
    }
  };

  const handleLogout = async () => {
    await logout();
    window.location.href = "/";
  };

  if (isLoading) {
    return (
      <div className="min-h-screen pt-24 pb-12 px-4 flex items-center justify-center">
        <div className="animate-spin w-12 h-12 border-4 border-[var(--primary)] border-t-transparent rounded-full"></div>
      </div>
    );
  }

  const user = profileData?.user;
  const stats = profileData?.stats || { total_matches: 0, wins: 0, losses: 0, win_rate: 0, skill_level: "Beginner" };

  return (
    <div className="min-h-screen pt-24 pb-12 px-4">
      <div className="max-w-4xl mx-auto">
        {error && <div className="mb-6 p-3 rounded-lg bg-[var(--error)]/10 text-[var(--error)] text-sm">{error}</div>}
        {success && <div className="mb-6 p-3 rounded-lg bg-[var(--success)]/10 text-[var(--success)] text-sm">{success}</div>}

        <div className="card mb-6">
          <div className="flex flex-col sm:flex-row items-center gap-6">
            <div className="w-24 h-24 rounded-full bg-gradient-to-br from-[var(--primary)] to-[var(--secondary)] flex items-center justify-center text-4xl text-white font-bold">
              {(user?.name || user?.username || "P").charAt(0).toUpperCase()}
            </div>

            <div className="flex-1 text-center sm:text-left">
              {isEditing ? (
                <div className="space-y-3">
                  <input
                    type="text"
                    className="input text-xl font-bold"
                    value={editForm.name}
                    onChange={(e) => setEditForm({ ...editForm, name: e.target.value })}
                    placeholder="Display Name"
                  />
                  <textarea
                    className="input resize-none"
                    rows={2}
                    value={editForm.bio}
                    onChange={(e) => setEditForm({ ...editForm, bio: e.target.value })}
                    placeholder="Bio..."
                  />
                </div>
              ) : (
                <>
                  <h1 className="text-2xl font-bold mb-1">{user?.name || user?.username}</h1>
                  <p className="text-[var(--muted)] mb-2">@{user?.username}</p>
                  <p>{user?.bio || "No bio yet"}</p>
                </>
              )}
            </div>

            <div className="flex gap-2">
              {isEditing ? (
                <>
                  <button onClick={() => setIsEditing(false)} className="btn-secondary py-2 px-4">Cancel</button>
                  <button onClick={handleSave} disabled={isSaving} className="btn-primary py-2 px-4">
                    {isSaving ? "Saving..." : "Save"}
                  </button>
                </>
              ) : (
                <>
                  <button onClick={() => setIsEditing(true)} className="btn-secondary py-2 px-4">Edit</button>
                  <button onClick={handleLogout} className="btn-secondary py-2 px-4 text-[var(--error)]">Logout</button>
                </>
              )}
            </div>
          </div>

          <div className="flex flex-wrap gap-3 mt-6 pt-6 border-t border-[var(--border)]">
            <span className="px-3 py-1 rounded-full text-sm font-medium bg-[var(--primary)]/10 text-[var(--primary)] capitalize">
              {user?.role}
            </span>
            <span className="px-3 py-1 rounded-full text-sm bg-[var(--surface)] text-[var(--muted)]">
              {stats.skill_level}
            </span>
          </div>
        </div>

        <div className="grid md:grid-cols-2 gap-6">
          <div className="card">
            <h2 className="text-lg font-semibold mb-4">Stats</h2>
            <div className="grid grid-cols-2 gap-4">
              <div className="p-3 rounded-lg bg-[var(--surface)]">
                <p className="text-[var(--muted)] text-sm">Matches</p>
                <p className="text-2xl font-bold">{stats.total_matches}</p>
              </div>
              <div className="p-3 rounded-lg bg-[var(--surface)]">
                <p className="text-[var(--muted)] text-sm">Win Rate</p>
                <p className="text-2xl font-bold text-[var(--success)]">{stats.win_rate?.toFixed(0)}%</p>
              </div>
              <div className="p-3 rounded-lg bg-[var(--surface)]">
                <p className="text-[var(--muted)] text-sm">Wins</p>
                <p className="text-2xl font-bold text-[var(--success)]">{stats.wins}</p>
              </div>
              <div className="p-3 rounded-lg bg-[var(--surface)]">
                <p className="text-[var(--muted)] text-sm">Losses</p>
                <p className="text-2xl font-bold text-[var(--error)]">{stats.losses}</p>
              </div>
            </div>
          </div>

          <div className="card">
            <h2 className="text-lg font-semibold mb-4">Win/Loss</h2>
            <div className="h-4 rounded-full overflow-hidden bg-[var(--surface)] flex">
              <div className="h-full bg-[var(--success)]" style={{ width: `${stats.win_rate || 0}%` }}></div>
              <div className="h-full bg-[var(--error)]" style={{ width: `${100 - (stats.win_rate || 0)}%` }}></div>
            </div>
            <div className="flex justify-between mt-2 text-sm">
              <span className="text-[var(--success)]">{stats.wins} Wins</span>
              <span className="text-[var(--error)]">{stats.losses} Losses</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

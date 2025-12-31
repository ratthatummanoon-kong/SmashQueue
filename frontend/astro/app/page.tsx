"use client";

import Link from "next/link"; 
import { useState, useEffect } from "react";
import { isAuthenticated } from "./lib/api";

export default function Home() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    const loggedIn = isAuthenticated();
    setIsLoggedIn(loggedIn);
    
    // Redirect authenticated users to dashboard
    if (loggedIn) {
      window.location.href = "/dashboard";
      return;
    }
  }, []);

  // If logged in, show loading while redirecting
  if (mounted && isLoggedIn) {
    return (
      <div className="min-h-screen pt-24 pb-12 px-4 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin w-12 h-12 border-4 border-[var(--primary)] border-t-transparent rounded-full mx-auto mb-4"></div>
          <p className="text-[var(--muted)]">Redirecting to dashboard...</p>
        </div>
      </div>
    );
  }

  // Guest landing page
  return (
    <div className="min-h-screen">
      {/* Hero Section */}
      <section className="relative pt-32 pb-20 px-4 overflow-hidden">
        {/* Background gradient */}
        <div className="absolute inset-0 bg-gradient-to-br from-[var(--primary)]/10 via-transparent to-[var(--secondary)]/10 pointer-events-none" />
        
        {/* Floating shuttlecock decoration */}
        <div className="absolute top-40 right-10 text-6xl opacity-20 animate-float hidden lg:block">
          🏸
        </div>
        <div className="absolute bottom-20 left-10 text-4xl opacity-15 animate-float hidden lg:block" style={{ animationDelay: "1s" }}>
          🏸
        </div>

        <div className="max-w-7xl mx-auto">
          <div className="text-center max-w-4xl mx-auto animate-fade-in">
            <span className="inline-block px-4 py-1.5 rounded-full bg-[var(--primary)]/10 text-[var(--primary)] text-sm font-medium mb-6">
              🏸 Badminton Queue Management
            </span>
            <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold leading-tight mb-6">
              Smarter Matches,{" "}
              <span className="gradient-text">Better Games</span>
            </h1>
            <p className="text-lg sm:text-xl text-[var(--muted)] mb-8 max-w-2xl mx-auto">
              SmashQueue streamlines matchmaking, queue management, and performance tracking 
              for your badminton social group. Say goodbye to chaotic court rotations.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Link href="/register" className="btn-primary text-lg px-8 py-3">
                Get Started Free
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
                </svg>
              </Link>
              <Link href="/login" className="btn-secondary text-lg px-8 py-3">
                Sign In
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 px-4 bg-[var(--surface)]">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-16">
            <h2 className="text-3xl sm:text-4xl font-bold mb-4">
              Everything You Need to Run Your <span className="gradient-text">Guan</span>
            </h2>
            <p className="text-[var(--muted)] text-lg max-w-2xl mx-auto">
              Powerful features designed specifically for badminton social groups
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {/* Feature 1*/}
            <div className="card hover:border-[var(--primary)] transition-colors group">
              <div className="w-12 h-12 rounded-xl bg-[var(--primary)]/10 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
                <svg className="w-6 h-6 text-[var(--primary)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                </svg>
              </div>
              <h3 className="text-xl font-semibold mb-2">Smart Matchmaking</h3>
              <p className="text-[var(--muted)]">
                Automatically balance teams based on skill levels for fair and competitive games every time.
              </p>
            </div>

            {/* Feature 2 */}
            <div className="card hover:border-[var(--secondary)] transition-colors group">
              <div className="w-12 h-12 rounded-xl bg-[var(--secondary)]/10 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
                <svg className="w-6 h-6 text-[var(--secondary)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                </svg>
              </div>
              <h3 className="text-xl font-semibold mb-2">Queue Management</h3>
              <p className="text-[var(--muted)]">
                Join the queue, track your position, and get notified when it&apos;s time to play.
              </p>
            </div>

            {/* Feature 3 */}
            <div className="card hover:border-[var(--accent)] transition-colors group">
              <div className="w-12 h-12 rounded-xl bg-[var(--accent)]/10 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
                <svg className="w-6 h-6 text-[var(--accent)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
              </div>
              <h3 className="text-xl font-semibold mb-2">Player Analytics</h3>
              <p className="text-[var(--muted)]">
                Track your win rate, skill progression, and performance trends over time.
              </p>
            </div>

            {/* Feature 4 */}
            <div className="card hover:border-[var(--success)] transition-colors group">
              <div className="w-12 h-12 rounded-xl bg-[var(--success)]/10 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
                <svg className="w-6 h-6 text-[var(--success)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <h3 className="text-xl font-semibold mb-2">Match History</h3>
              <p className="text-[var(--muted)]">
                Complete record of all your games with detailed scores and statistics.
              </p>
            </div>

            {/* Feature 5 */}
            <div className="card hover:border-[var(--warning)] transition-colors group">
              <div className="w-12 h-12 rounded-xl bg-[var(--warning)]/10 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
                <svg className="w-6 h-6 text-[var(--warning)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
              </div>
              <h3 className="text-xl font-semibold mb-2">Secure & Private</h3>
              <p className="text-[var(--muted)]">
                Your data is protected with industry-standard encryption and authentication.
              </p>
            </div>

            {/* Feature 6 */}
            <div className="card hover:border-[var(--error)] transition-colors group">
              <div className="w-12 h-12 rounded-xl bg-[var(--error)]/10 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
                <svg className="w-6 h-6 text-[var(--error)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5.121 17.804A13.937 13.937 0 0112 16c2.5 0 4.847.655 6.879 1.804M15 10a3 3 0 11-6 0 3 3 0 016 0zm6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <h3 className="text-xl font-semibold mb-2">Role Management</h3>
              <p className="text-[var(--muted)]">
                Assign organizers to manage queues and matches for your group.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 px-4">
        <div className="max-w-3xl mx-auto text-center">
          <h2 className="text-3xl sm:text-4xl font-bold mb-4">
            Ready to Level Up Your <span className="gradient-text">Guan</span>?
          </h2>
          <p className="text-[var(--muted)] text-lg mb-8">
            Join SmashQueue today and experience organized, fair, and fun badminton sessions.
          </p>
          <Link href="/register" className="btn-primary text-lg px-8 py-3 inline-flex items-center gap-2">
            Start Playing Now
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
            </svg>
          </Link>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-[var(--border)] py-8 px-4">
        <div className="max-w-7xl mx-auto text-center text-[var(--muted)] text-sm">
          <p>© 2025 SmashQueue. Made with 🏸 for badminton lovers.</p>
        </div>
      </footer>
    </div>
  );
}

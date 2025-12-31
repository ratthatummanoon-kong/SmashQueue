"use client";

import Link from "next/link";
import { useRouter, usePathname } from "next/navigation";
import { useState, useEffect } from "react";
import { getAccessToken, getUser, logout, isAdmin, User } from "../lib/api";

export default function Navbar() {
  const router = useRouter();
  const pathname = usePathname();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [user, setUser] = useState<User | null>(null);
  const [mounted, setMounted] = useState(false);

  // Re-check auth state when pathname changes
  useEffect(() => {
    setMounted(true);
    checkAuth();
  }, [pathname]);

  const checkAuth = () => {
    const token = getAccessToken();
    if (token) {
      const userData = getUser();
      setUser(userData);
    } else {
      setUser(null);
    }
  };

  const handleLogout = async () => {
    await logout();
    setUser(null);
    setIsMenuOpen(false);
    router.push("/");
    router.refresh();
  };

  // Don't render auth buttons until mounted to prevent hydration mismatch
  if (!mounted) {
    return (
      <nav className="fixed top-0 left-0 right-0 z-50 glass">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <Link href="/" className="flex items-center gap-2">
              <span className="text-2xl">🏸</span>
              <span className="text-xl font-bold gradient-text">SmashQueue</span>
            </Link>
          </div>
        </div>
      </nav>
    );
  }

  const isLoggedIn = user !== null;
  const showAdminLink = isLoggedIn && isAdmin();

  return (
    <nav className="fixed top-0 left-0 right-0 z-50 glass">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <Link href="/" className="flex items-center gap-2 group">
            <span className="text-2xl">🏸</span>
            <span className="text-xl font-bold gradient-text">SmashQueue</span>
          </Link>

          {/* Desktop Navigation */}
          <div className="hidden md:flex items-center gap-6">
            <Link
              href="/"
              className="text-[var(--muted)] hover:text-[var(--foreground)] font-medium transition-colors"
            >
              Home
            </Link>
            {isLoggedIn && (
              <>
                <Link
                  href="/dashboard"
                  className="text-[var(--muted)] hover:text-[var(--foreground)] font-medium transition-colors"
                >
                  Dashboard
                </Link>
                <Link
                  href="/profile"
                  className="text-[var(--muted)] hover:text-[var(--foreground)] font-medium transition-colors"
                >
                  Profile
                </Link>
                {showAdminLink && (
                  <Link
                    href="/admin"
                    className="text-[var(--accent)] hover:text-[var(--foreground)] font-medium transition-colors"
                  >
                    Admin
                  </Link>
                )}
              </>
            )}
            <div className="flex items-center gap-3 ml-4">
              {isLoggedIn ? (
                <>
                  <span className="text-sm text-[var(--muted)]">
                    👤 {user?.name || user?.username}
                    {user?.role !== 'player' && (
                      <span className="ml-1 text-xs text-[var(--accent)] capitalize">
                        ({user?.role})
                      </span>
                    )}
                  </span>
                  <button
                    onClick={handleLogout}
                    className="btn-secondary text-sm py-2 px-4"
                  >
                    Logout
                  </button>
                </>
              ) : (
                <>
                  <Link href="/login" className="btn-secondary text-sm py-2 px-4">
                    Login
                  </Link>
                  <Link href="/register" className="btn-primary text-sm py-2 px-4">
                    Sign Up
                  </Link>
                </>
              )}
            </div>
          </div>

          {/* Mobile Menu Button */}
          <button
            onClick={() => setIsMenuOpen(!isMenuOpen)}
            className="md:hidden p-2 rounded-lg hover:bg-[var(--surface)] transition-colors"
            aria-label="Toggle menu"
          >
            <svg
              className="w-6 h-6"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              {isMenuOpen ? (
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              ) : (
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 6h16M4 12h16M4 18h16"
                />
              )}
            </svg>
          </button>
        </div>

        {/* Mobile Menu */}
        {isMenuOpen && (
          <div className="md:hidden py-4 border-t border-[var(--border)]">
            <div className="flex flex-col gap-3">
              <Link
                href="/"
                className="px-3 py-2 rounded-lg hover:bg-[var(--surface)] transition-colors"
                onClick={() => setIsMenuOpen(false)}
              >
                Home
              </Link>
              {isLoggedIn && (
                <>
                  <Link
                    href="/dashboard"
                    className="px-3 py-2 rounded-lg hover:bg-[var(--surface)] transition-colors"
                    onClick={() => setIsMenuOpen(false)}
                  >
                    Dashboard
                  </Link>
                  <Link
                    href="/profile"
                    className="px-3 py-2 rounded-lg hover:bg-[var(--surface)] transition-colors"
                    onClick={() => setIsMenuOpen(false)}
                  >
                    Profile
                  </Link>
                  {showAdminLink && (
                    <Link
                      href="/admin"
                      className="px-3 py-2 rounded-lg hover:bg-[var(--surface)] transition-colors text-[var(--accent)]"
                      onClick={() => setIsMenuOpen(false)}
                    >
                      Admin Panel
                    </Link>
                  )}
                </>
              )}
              <div className="flex flex-col gap-2 mt-2 pt-3 border-t border-[var(--border)]">
                {isLoggedIn ? (
                  <>
                    <div className="px-3 py-2 text-sm text-[var(--muted)]">
                      Logged in as <strong>{user?.name || user?.username}</strong>
                      {user?.role !== 'player' && (
                        <span className="ml-1 text-[var(--accent)] capitalize">
                          ({user?.role})
                        </span>
                      )}
                    </div>
                    <button
                      onClick={handleLogout}
                      className="btn-secondary text-center"
                    >
                      Logout
                    </button>
                  </>
                ) : (
                  <>
                    <Link
                      href="/login"
                      className="btn-secondary text-center"
                      onClick={() => setIsMenuOpen(false)}
                    >
                      Login
                    </Link>
                    <Link
                      href="/register"
                      className="btn-primary text-center"
                      onClick={() => setIsMenuOpen(false)}
                    >
                      Sign Up
                    </Link>
                  </>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </nav>
  );
}

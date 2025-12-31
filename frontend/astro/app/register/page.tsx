"use client";

import Link from "next/link";
import { useState, useMemo, useEffect } from "react";
import { register, getAccessToken } from "../lib/api";

export default function RegisterPage() {
  const [formData, setFormData] = useState({
    username: "",
    password: "",
    confirmPassword: "",
    phone: "",
    name: "",
  });
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");

  // Redirect if already logged in
  useEffect(() => {
    if (getAccessToken()) {
      window.location.href = "/dashboard";
    }
  }, []);

  const passwordValidation = useMemo(() => {
    const password = formData.password;
    return {
      minLength: password.length >= 8,
      hasLowercase: /[a-z]/.test(password),
      hasUppercase: /[A-Z]/.test(password),
      hasNumber: /[0-9]/.test(password),
      hasSpecial: /[!@#$%^&*(),.?":{}|<>]/.test(password),
      notUsername: formData.username.length === 0 || password !== formData.username,
    };
  }, [formData.password, formData.username]);

  const isPasswordValid = useMemo(() => {
    return Object.values(passwordValidation).every(Boolean);
  }, [passwordValidation]);

  const passwordsMatch = formData.password === formData.confirmPassword;
  const isPhoneValid = /^0[89]\d{8}$/.test(formData.phone);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!isPasswordValid) {
      setError("Please meet all password requirements");
      return;
    }

    if (!passwordsMatch) {
      setError("Passwords do not match");
      return;
    }

    if (!isPhoneValid) {
      setError("Please enter a valid Thai phone number (08x or 09x, 10 digits)");
      return;
    }

    setIsLoading(true);

    try {
      const response = await register(
        formData.username,
        formData.password,
        formData.confirmPassword,
        formData.phone,
        formData.name
      );

      if (response.success && response.data) {
        window.location.href = "/dashboard";
      } else {
        setError(response.error?.message || "Registration failed.");
      }
    } catch {
      setError("Network error. Please check your connection.");
    } finally {
      setIsLoading(false);
    }
  };

  const ValidationItem = ({ valid, text }: { valid: boolean; text: string }) => (
    <div className={`flex items-center gap-2 text-sm ${valid ? "text-[var(--success)]" : "text-[var(--muted)]"}`}>
      {valid ? "✓" : "○"} {text}
    </div>
  );

  const EyeIcon = ({ show, onClick }: { show: boolean; onClick: () => void }) => (
    <button
      type="button"
      onClick={onClick}
      className="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--muted)] hover:text-[var(--foreground)] transition-colors p-1"
      tabIndex={-1}
    >
      {show ? (
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
        </svg>
      ) : (
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
        </svg>
      )}
    </button>
  );

  return (
    <div className="min-h-screen pt-24 pb-12 px-4 flex items-center justify-center">
      <div className="fixed inset-0 bg-gradient-to-br from-[var(--primary)]/5 via-transparent to-[var(--secondary)]/5 pointer-events-none" />

      <div className="w-full max-w-md relative">
        <div className="card animate-fade-in">
          <div className="text-center mb-8">
            <div className="inline-block p-3 rounded-full bg-[var(--primary)]/10 mb-4">
              <span className="text-3xl">🏸</span>
            </div>
            <h1 className="text-2xl font-bold mb-2">Create Account</h1>
            <p className="text-[var(--muted)]">Join SmashQueue today</p>
          </div>

          {error && (
            <div className="mb-6 p-3 rounded-lg bg-[var(--error)]/10 border border-[var(--error)]/20 text-[var(--error)] text-sm">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="name" className="label">Display Name</label>
              <input
                type="text"
                id="name"
                className="input"
                placeholder="Your name (e.g., Somchai)"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              />
            </div>

            <div>
              <label htmlFor="phone" className="label">Phone Number *</label>
              <input
                type="tel"
                id="phone"
                className="input"
                placeholder="08xxxxxxxx or 09xxxxxxxx"
                value={formData.phone}
                onChange={(e) => {
                  const val = e.target.value.replace(/\D/g, "").slice(0, 10);
                  setFormData({ ...formData, phone: val });
                }}
                required
              />
              {formData.phone && (
                <div className="mt-2">
                  <ValidationItem 
                    valid={isPhoneValid} 
                    text={isPhoneValid ? "Valid phone number" : "Must be 10 digits starting with 08 or 09"} 
                  />
                </div>
              )}
            </div>

            <div>
              <label htmlFor="username" className="label">Username *</label>
              <input
                type="text"
                id="username"
                className="input"
                placeholder="Choose a username"
                value={formData.username}
                onChange={(e) => setFormData({ ...formData, username: e.target.value.toLowerCase().replace(/\s/g, "") })}
                required
              />
            </div>

            <div>
              <label htmlFor="password" className="label">Password *</label>
              <div className="relative">
                <input
                  type={showPassword ? "text" : "password"}
                  id="password"
                  className="input pr-12"
                  placeholder="Create a strong password"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  required
                />
                <EyeIcon show={showPassword} onClick={() => setShowPassword(!showPassword)} />
              </div>
              {formData.password && (
                <div className="mt-3 p-3 rounded-lg bg-[var(--surface)] space-y-1.5">
                  <ValidationItem valid={passwordValidation.minLength} text="At least 8 characters" />
                  <ValidationItem valid={passwordValidation.hasLowercase} text="One lowercase (a-z)" />
                  <ValidationItem valid={passwordValidation.hasUppercase} text="One uppercase (A-Z)" />
                  <ValidationItem valid={passwordValidation.hasNumber} text="One number (0-9)" />
                  <ValidationItem valid={passwordValidation.hasSpecial} text="One special character" />
                  <ValidationItem valid={passwordValidation.notUsername} text="Not same as username" />
                </div>
              )}
            </div>

            <div>
              <label htmlFor="confirmPassword" className="label">Confirm Password *</label>
              <div className="relative">
                <input
                  type={showConfirmPassword ? "text" : "password"}
                  id="confirmPassword"
                  className="input pr-12"
                  placeholder="Confirm your password"
                  value={formData.confirmPassword}
                  onChange={(e) => setFormData({ ...formData, confirmPassword: e.target.value })}
                  required
                />
                <EyeIcon show={showConfirmPassword} onClick={() => setShowConfirmPassword(!showConfirmPassword)} />
              </div>
              {formData.confirmPassword && (
                <div className="mt-2">
                  <ValidationItem valid={passwordsMatch} text={passwordsMatch ? "Passwords match" : "Passwords do not match"} />
                </div>
              )}
            </div>

            <button
              type="submit"
              disabled={isLoading || !isPasswordValid || !passwordsMatch || !isPhoneValid}
              className="btn-primary w-full py-3 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? "Creating account..." : "Create Account"}
            </button>
          </form>

          <div className="relative my-8">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-[var(--border)]"></div>
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="px-2 bg-[var(--surface-elevated)] text-[var(--muted)]">
                Already have an account?
              </span>
            </div>
          </div>

          <Link href="/login" className="btn-secondary w-full py-3 text-center block">
            Sign In
          </Link>
        </div>
      </div>
    </div>
  );
}

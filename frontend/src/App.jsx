import { useEffect, useState } from "react";
import { BrowserRouter, Navigate, Route, Routes, useLocation } from "react-router-dom";
import { api } from "./lib/api";
import { ProtectedLayout, CustomerLayout, OwnerLayout } from "./layouts";
import AuthPage, { VerifyEmail } from "./pages/auth";
import { Home, Booking, History, Membership as CustomerMembership, Notifications } from "./pages/customer";
import { Overview, Bookings, Rooms, Equipment, Membership as OwnerMembership, SettingsPage } from "./pages/owner";
import "./styles.css";

function GuestAuth({ onLogin }) { const { pathname } = useLocation(); return <AuthPage mode={pathname === "/reset-password" ? "reset" : authMode(pathname)} onLogin={onLogin} />; }

function authMode(path) { return path === "/register" ? "register" : path === "/forgot-password" ? "forgot" : "login"; }

export default function App() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  useEffect(() => { api("/api/auth/me").then((data) => setUser(data.user)).catch(() => {}).finally(() => setLoading(false)); }, []);
  if (loading) return <div className="loading-screen"><span className="brand-mark">B</span><span>正在打开 BandRoom…</span></div>;
  return <BrowserRouter><Routes>
    <Route path="/verify-email" element={<VerifyEmail />} />
    <Route path="/reset-password" element={<GuestAuth onLogin={setUser} />} />
    {!user ? <Route path="*" element={<GuestAuth onLogin={setUser} />} /> : <>
      <Route element={<ProtectedLayout user={user} onLogout={() => setUser(null)} />}>
        {user.role === "customer" ? <Route path="/customer" element={<CustomerLayout />}><Route index element={<Navigate to="home" replace />} /><Route path="home" element={<Home />} /><Route path="bookings" element={<Booking />} /><Route path="history" element={<History />} /><Route path="membership" element={<CustomerMembership />} /><Route path="notifications" element={<Notifications />} /></Route> : <Route path="/owner" element={<OwnerLayout />}><Route index element={<Overview />} /><Route path="bookings" element={<Bookings />} /><Route path="rooms" element={<Rooms />} /><Route path="equipment" element={<Equipment />} /><Route path="membership" element={<OwnerMembership />} /><Route path="settings" element={<SettingsPage />} /></Route>}
      </Route>
      <Route path="*" element={<Navigate to={user.role === "owner" ? "/owner" : "/customer/bookings"} replace />} />
    </>}
  </Routes></BrowserRouter>;
}

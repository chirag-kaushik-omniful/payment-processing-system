import { NavLink, Outlet } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import "./Layout.css";

const nav = [
  { to: "/", label: "Dashboard", end: true },
  { to: "/payments", label: "Payments" },
  { to: "/wallet", label: "Wallet" },
  { to: "/payment-methods", label: "Payment methods" },
  { to: "/admin", label: "Admin" },
];

export function Layout() {
  const { user, logout } = useAuth();

  return (
    <div className="layout">
      <aside className="sidebar">
        <div className="brand">
          <span className="brand-icon">◆</span>
          <span>Payment Platform</span>
        </div>
        <nav className="nav">
          {nav.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) => (isActive ? "nav-link active" : "nav-link")}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
        <div className="sidebar-footer">
          <p className="user-email">{user?.email}</p>
          <button type="button" className="btn btn-secondary btn-sm" onClick={logout}>
            Sign out
          </button>
        </div>
      </aside>
      <main className="main">
        <Outlet />
      </main>
    </div>
  );
}

import { Navigate, Route, Routes } from "react-router-dom";
import { Layout } from "./components/Layout";
import { useAuth } from "./context/AuthContext";
import { Admin } from "./pages/Admin";
import { Dashboard } from "./pages/Dashboard";
import { Login } from "./pages/Login";
import { PaymentMethods } from "./pages/PaymentMethods";
import { Payments } from "./pages/Payments";
import { Signup } from "./pages/Signup";
import { WalletPage } from "./pages/Wallet";

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  return <>{children}</>;
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/signup" element={<Signup />} />
      <Route
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route path="/" element={<Dashboard />} />
        <Route path="/payments" element={<Payments />} />
        <Route path="/wallet" element={<WalletPage />} />
        <Route path="/payment-methods" element={<PaymentMethods />} />
        <Route path="/admin" element={<Admin />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

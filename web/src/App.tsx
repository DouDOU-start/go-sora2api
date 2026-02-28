import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './store/authStore'
import Layout from './components/Layout'
import ToastContainer from './components/ui/Toast'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import AccountList from './pages/AccountList'
import GroupList from './pages/GroupList'
import APIKeyList from './pages/APIKeyList'
import TaskList from './pages/TaskList'
import TaskDetail from './pages/TaskDetail'
import Settings from './pages/Settings'
import CharacterList from './pages/CharacterList'
import Docs from './pages/Docs'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { token } = useAuthStore()
  if (!token) return <Navigate to="/login" replace />
  return <>{children}</>
}

// 仅管理员可访问的路由保护
function AdminRoute({ children }: { children: React.ReactNode }) {
  const { role } = useAuthStore()
  if (role !== 'admin') return <Navigate to="/characters" replace />
  return <>{children}</>
}

// viewer 的默认首页
function DefaultRedirect() {
  const { role } = useAuthStore()
  if (role === 'admin') return <Dashboard />
  return <Navigate to="/characters" replace />
}

export default function App() {
  return (
    <BrowserRouter>
      <ToastContainer />
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route
          element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }
        >
          <Route path="/" element={<DefaultRedirect />} />
          <Route path="/accounts" element={<AdminRoute><AccountList /></AdminRoute>} />
          <Route path="/groups" element={<AdminRoute><GroupList /></AdminRoute>} />
          <Route path="/api-keys" element={<AdminRoute><APIKeyList /></AdminRoute>} />
          <Route path="/tasks" element={<TaskList />} />
          <Route path="/tasks/:id" element={<TaskDetail />} />
          <Route path="/characters" element={<CharacterList />} />
          <Route path="/settings" element={<AdminRoute><Settings /></AdminRoute>} />
          <Route path="/docs" element={<Docs />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

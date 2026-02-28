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
          <Route path="/" element={<Dashboard />} />
          <Route path="/accounts" element={<AccountList />} />
          <Route path="/groups" element={<GroupList />} />
          <Route path="/api-keys" element={<APIKeyList />} />
          <Route path="/tasks" element={<TaskList />} />
          <Route path="/tasks/:id" element={<TaskDetail />} />
          <Route path="/characters" element={<CharacterList />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="/docs" element={<Docs />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

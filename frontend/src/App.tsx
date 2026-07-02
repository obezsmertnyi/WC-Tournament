import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import Layout from './components/Layout'
import Calendar from './pages/Calendar'
import Groups from './pages/Groups'
import GroupDetail from './pages/GroupDetail'
import Competition from './pages/Competition'
import MyHistory from './pages/MyHistory'
import Audit from './pages/Audit'
import Profile from './pages/Profile'
import Bracket from './pages/Bracket'
import AI from './pages/AI'
import ComingSoon from './pages/ComingSoon'
import { AuthProvider } from './auth/AuthContext'
import { PredictionsProvider } from './predictions/PredictionsContext'

/**
 * Providers live *inside* the router so descendants (the app bar, profile page)
 * can use router hooks like `useNavigate`. Predictions depend on auth, so the
 * predictions provider nests inside the auth provider.
 */
function Providers() {
  return (
    <AuthProvider>
      <PredictionsProvider>
        <Layout />
      </PredictionsProvider>
    </AuthProvider>
  )
}

const router = createBrowserRouter([
  {
    element: <Providers />,
    children: [
      { index: true, element: <Calendar /> },
      { path: 'groups', element: <Groups /> },
      { path: 'groups/:letter', element: <GroupDetail /> },
      { path: 'competition', element: <Competition /> },
      { path: 'history', element: <MyHistory /> },
      { path: 'audit', element: <Audit /> },
      { path: 'profile', element: <Profile /> },
      { path: 'bracket', element: <Bracket /> },
      { path: 'ai', element: <AI /> },
      { path: '*', element: <ComingSoon section="notFound" /> },
    ],
  },
])

export default function App() {
  return <RouterProvider router={router} />
}

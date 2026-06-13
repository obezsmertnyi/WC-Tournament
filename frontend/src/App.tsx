import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import Layout from './components/Layout'
import Calendar from './pages/Calendar'
import Groups from './pages/Groups'
import GroupDetail from './pages/GroupDetail'
import ComingSoon from './pages/ComingSoon'

const router = createBrowserRouter([
  {
    element: <Layout />,
    children: [
      { index: true, element: <Calendar /> },
      { path: 'groups', element: <Groups /> },
      { path: 'groups/:letter', element: <GroupDetail /> },
      {
        path: 'leaderboard',
        element: <ComingSoon section="leaderboard" />,
      },
      {
        path: 'bracket',
        element: <ComingSoon section="bracket" />,
      },
      {
        path: '*',
        element: <ComingSoon section="notFound" />,
      },
    ],
  },
])

export default function App() {
  return <RouterProvider router={router} />
}

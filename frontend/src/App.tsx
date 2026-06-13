import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import Layout from './components/Layout'
import Calendar from './pages/Calendar'
import ComingSoon from './pages/ComingSoon'

const router = createBrowserRouter([
  {
    element: <Layout />,
    children: [
      { index: true, element: <Calendar /> },
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

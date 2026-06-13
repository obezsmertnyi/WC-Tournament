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
        element: (
          <ComingSoon
            title="Лідери"
            description="Таблиця лідерів зʼявиться, щойно почнуться матчі та надійдуть перші прогнози."
          />
        ),
      },
      {
        path: 'bracket',
        element: (
          <ComingSoon
            title="Турнірна сітка"
            description="Інтерактивна сітка плей-оф відкриється після завершення групового етапу."
          />
        ),
      },
      {
        path: '*',
        element: (
          <ComingSoon
            title="Сторінку не знайдено"
            description="Можливо, посилання застаріло. Поверніться до календаря."
          />
        ),
      },
    ],
  },
])

export default function App() {
  return <RouterProvider router={router} />
}

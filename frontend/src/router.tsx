import { createBrowserRouter, Outlet } from 'react-router-dom';
import RootLayout from './layouts/RootLayout';
import TalkPage from './pages/TalkPage';
import HistoryListPage from './pages/HistoryListPage';
import HistoryDetailPage from './pages/HistoryDetailPage';
import PersonaManagePage from './pages/PersonaManagePage';
import PersonaEditor from './components/PersonaEditor';

export const router = createBrowserRouter([
  {
    path: '/',
    element: <RootLayout />,
    children: [
      { index: true, element: <TalkPage /> },
      { path: 'history', element: <HistoryListPage /> },
      { path: 'history/:id', element: <HistoryDetailPage /> },
      { path: 'personas', element: <PersonaManagePage /> },
      { path: 'personas/new', element: <PersonaEditor /> },
      { path: 'personas/:id/edit', element: <PersonaEditor /> },
    ],
  },
]);

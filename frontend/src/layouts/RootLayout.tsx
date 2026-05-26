import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useEffect } from 'react';
import Header from '../components/Header';
import NavSidebar from '../components/NavSidebar';
import { usePersonaStore } from '../stores/usePersonaStore';

export default function RootLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { fetchPersonas } = usePersonaStore();

  // Determine current page and header title
  const path = location.pathname;
  const isPersonasPage = path === '/personas' || path.startsWith('/personas/');
  const isHistoryPage = path === '/history' || path.startsWith('/history/');
  const isTalkPage = path === '/' || path.startsWith('/?');

  const navPage = isPersonasPage ? 'personas' : 'talk';

  let headerTitle = '';
  if (isPersonasPage) headerTitle = 'tabPersonas';
  else if (isHistoryPage) headerTitle = 'pageHistory';
  else headerTitle = 'pageRoundtable';

  // Load personas on mount
  useEffect(() => {
    fetchPersonas();
  }, [fetchPersonas]);

  const handleNavigate = (page: 'talk' | 'personas') => {
    if (page === 'talk') {
      navigate('/');
    } else {
      navigate('/personas');
    }
  };

  // Persona editor pages have their own layout
  if (path.startsWith('/personas/') && path !== '/personas') {
    return (
      <div className="h-screen flex flex-col bg-white">
        <Outlet />
      </div>
    );
  }

  // History pages have their own layout (no sidebar)
  if (isHistoryPage) {
    return (
      <div className="min-h-screen bg-white">
        <Header title={headerTitle} showNav={true} />
        <Outlet />
      </div>
    );
  }

  // Talk and Personas pages share sidebar layout
  return (
    <div className="h-screen flex flex-col">
      <Header title={headerTitle} showNav={true} />
      <div className="flex flex-1 overflow-hidden">
        <NavSidebar active={navPage} onNavigate={handleNavigate} />
        <main className="flex-1 overflow-hidden flex">
          <Outlet />
        </main>
      </div>
    </div>
  );
}

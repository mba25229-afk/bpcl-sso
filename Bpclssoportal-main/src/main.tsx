import { createRoot } from 'react-dom/client';
import { useState } from 'react';
import App from './app/App.tsx';
import { LoginPage } from './app/LoginPage.tsx';
import './styles/index.css';

function Root() {
  const [authed, setAuthed] = useState(() => !!localStorage.getItem('bpcl_token'));

  if (!authed) {
    return <LoginPage onLogin={() => setAuthed(true)} />;
  }
  return <App onLogout={() => setAuthed(false)} />;
}

createRoot(document.getElementById('root')!).render(<Root />);

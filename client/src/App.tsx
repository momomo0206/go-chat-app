import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import LoginPage from './pages/LoginPage';
import { AuthProvider } from './context/AuthContext';
import { ToastProvider } from './context/ToastContext';
import SignupPage from './pages/SignupPage';
import RoomsPage from './pages/RoomPage';
import Protected from './components/Protected';
import AboutPage from './pages/AboutPage';
import { ToastContainer } from './components/Toast';
import ChatPage from './pages/ChatPage';
import ProfilePage from './pages/ProfilePage';
import PrivacyPage from './pages/PrivacyPage';
import TermsPage from './pages/TermsPage';

export default function App() {
  return (
    <ToastProvider>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route path='/login' element={<LoginPage />} />
            <Route path='/signup' element={<SignupPage />} />
            <Route path='/about' element={<AboutPage />} />
            <Route path='/privacy' element={<PrivacyPage />} />
            <Route path='/terms' element={<TermsPage />} />

            <Route element={<Protected />}>
              <Route path='/rooms' element={<RoomsPage />} />
              <Route path='/chat/:roomId' element={<ChatPage />} />
              <Route path='/profile' element={<ProfilePage />} />
              <Route index element={<Navigate to='/rooms' replace />} />
            </Route>

            <Route path='*' element={<Navigate to='/login' replace />} />
          </Routes>
          <ToastContainer />
        </BrowserRouter>
      </AuthProvider>
    </ToastProvider>
  );
}

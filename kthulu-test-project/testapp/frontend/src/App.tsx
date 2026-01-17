// @kthulu:frontend:app
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import AdminLayout from './components/layouts/AdminLayout';

// Module Imports


const App = () => {
  return (
    <Router>
      <AdminLayout>
        <Routes>
          <Route path="/" element={<div className="bg-white p-6 rounded-xl shadow-sm border border-slate-100"><h1 className="text-2xl font-bold text-slate-800 mb-2">Welcome Back</h1><p className="text-slate-600">This is your new Kthulu Dashboard.</p></div>} />
          
          <Route path="/settings" element={<div className="bg-white p-6 rounded-xl shadow-sm border border-slate-100"><h1 className="text-2xl font-bold text-slate-800">Settings</h1></div>} />
        </Routes>
      </AdminLayout>
    </Router>
  );
};

export default App;

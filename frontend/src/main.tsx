import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { VehicleList } from './pages/VehicleList'
import { VehicleDetail } from './pages/VehicleDetail'
import { AddFillup } from './pages/AddFillup'
import './index.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<VehicleList />} />
        <Route path="/vehicles/:id" element={<VehicleDetail />} />
        <Route path="/vehicles/:id/fillup" element={<AddFillup />} />
      </Routes>
    </BrowserRouter>
  </React.StrictMode>
)

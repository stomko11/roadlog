import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api'

interface Vehicle { id: number; name: string; make: string; model: string; year: number; plate: string }

export function VehicleList() {
  const [vehicles, setVehicles] = useState<Vehicle[]>([])
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ name: '', make: '', model: '', year: new Date().getFullYear(), plate: '' })
  const nav = useNavigate()

  useEffect(() => { api<Vehicle[]>('/vehicles').then(setVehicles) }, [])

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    const v = await api<Vehicle>('/vehicles', { method: 'POST', body: JSON.stringify(form) })
    setVehicles([v, ...vehicles])
    setShowForm(false)
    setForm({ name: '', make: '', model: '', year: new Date().getFullYear(), plate: '' })
  }

  return (
    <div className="container">
      <div className="header">
        <h1>🚗 Roadlog</h1>
        <button className="btn btn-primary" onClick={() => setShowForm(!showForm)}>+ Vehicle</button>
      </div>

      {showForm && (
        <form onSubmit={submit} className="card">
          <div className="form-group"><label>Name</label><input required value={form.name} onChange={e => setForm({...form, name: e.target.value})} placeholder="e.g. Daily Driver" /></div>
          <div className="form-group"><label>Make</label><input value={form.make} onChange={e => setForm({...form, make: e.target.value})} placeholder="e.g. Volkswagen" /></div>
          <div className="form-group"><label>Model</label><input value={form.model} onChange={e => setForm({...form, model: e.target.value})} placeholder="e.g. Golf" /></div>
          <div className="form-group"><label>Year</label><input type="number" value={form.year} onChange={e => setForm({...form, year: +e.target.value})} /></div>
          <div className="form-group"><label>Plate</label><input value={form.plate} onChange={e => setForm({...form, plate: e.target.value})} /></div>
          <button className="btn btn-primary btn-block" type="submit">Add Vehicle</button>
        </form>
      )}

      {vehicles.length === 0 && !showForm && <p className="empty">No vehicles yet. Add one to get started.</p>}

      {vehicles.map(v => (
        <div key={v.id} className="card" onClick={() => nav(`/vehicles/${v.id}`)}>
          <strong>{v.name}</strong>
          <div style={{color: 'var(--muted)', fontSize: '0.85rem'}}>{v.year} {v.make} {v.model} · {v.plate}</div>
        </div>
      ))}
    </div>
  )
}

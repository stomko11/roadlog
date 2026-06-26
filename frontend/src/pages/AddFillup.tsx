import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { api } from '../api'

interface Prefill { odometer: number | null; pricePerUnit: number | null; station: string | null; fullTank: boolean | null }

export function AddFillup() {
  const { id } = useParams()
  const nav = useNavigate()
  const [prefill, setPrefill] = useState<Prefill | null>(null)
  const [form, setForm] = useState({ date: new Date().toISOString().slice(0, 10), odometer: '', fuelAmount: '', pricePerUnit: '', totalCost: '', station: '', fullTank: true, notes: '' })

  useEffect(() => {
    api<Prefill>(`/vehicles/${id}/fillups/prefill`).then(p => {
      setPrefill(p)
      setForm(f => ({
        ...f,
        odometer: p.odometer ? String(p.odometer) : '',
        pricePerUnit: p.pricePerUnit ? String(p.pricePerUnit) : '',
        station: p.station || '',
        fullTank: p.fullTank ?? true,
      }))
    })
  }, [id])

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    const body = {
      date: new Date(form.date).toISOString(),
      odometer: +form.odometer,
      fuelAmount: +form.fuelAmount,
      pricePerUnit: +form.pricePerUnit,
      totalCost: +form.totalCost || (+form.fuelAmount * +form.pricePerUnit),
      station: form.station,
      fullTank: form.fullTank,
      notes: form.notes,
    }
    await api(`/vehicles/${id}/fillups`, { method: 'POST', body: JSON.stringify(body) })
    nav(`/vehicles/${id}`)
  }

  const hint = (field: string) => prefill ? <div className="prefill-hint">Pre-filled from last entry</div> : null

  return (
    <div className="container">
      <div className="header"><a className="back" onClick={() => nav(-1)}>← Cancel</a><h1>New Fill-up</h1></div>
      <form onSubmit={submit}>
        <div className="form-group"><label>Date</label><input type="date" required value={form.date} onChange={e => setForm({...form, date: e.target.value})} /></div>

        <div className="form-group">
          <label>Odometer (km)</label>
          <input type="number" step="1" required value={form.odometer} onChange={e => setForm({...form, odometer: e.target.value})} placeholder="Current reading" />
          {prefill?.odometer && hint('odometer')}
        </div>

        <div className="form-group">
          <label>Fuel Amount (L)</label>
          <input type="number" step="0.01" required value={form.fuelAmount} onChange={e => setForm({...form, fuelAmount: e.target.value})} placeholder="Liters filled" />
        </div>

        <div className="form-group">
          <label>Price per Liter (€)</label>
          <input type="number" step="0.001" required value={form.pricePerUnit} onChange={e => setForm({...form, pricePerUnit: e.target.value})} placeholder="e.g. 1.459" />
          {prefill?.pricePerUnit && hint('pricePerUnit')}
        </div>

        <div className="form-group">
          <label>Total Cost (€) — auto-calculated if empty</label>
          <input type="number" step="0.01" value={form.totalCost} onChange={e => setForm({...form, totalCost: e.target.value})} placeholder={form.fuelAmount && form.pricePerUnit ? `${(+form.fuelAmount * +form.pricePerUnit).toFixed(2)}` : ''} />
        </div>

        <div className="form-group">
          <label>Station</label>
          <input value={form.station} onChange={e => setForm({...form, station: e.target.value})} placeholder="e.g. Shell, OMV..." />
          {prefill?.station && hint('station')}
        </div>

        <div className="form-group" style={{display: 'flex', alignItems: 'center', gap: '0.5rem'}}>
          <input type="checkbox" id="fullTank" checked={form.fullTank} onChange={e => setForm({...form, fullTank: e.target.checked})} />
          <label htmlFor="fullTank" style={{margin: 0}}>Full tank</label>
        </div>

        <div className="form-group"><label>Notes</label><input value={form.notes} onChange={e => setForm({...form, notes: e.target.value})} placeholder="Optional" /></div>

        <button className="btn btn-primary btn-block" type="submit">Save Fill-up</button>
      </form>
    </div>
  )
}

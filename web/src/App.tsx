import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import BoardList from './pages/BoardList'
import BoardView from './pages/BoardView'

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<BoardList />} />
        <Route path="/boards/:boardId" element={<BoardView />} />
      </Route>
    </Routes>
  )
}

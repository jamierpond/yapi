import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'

function HelloWorld() {
  return (
    <div>
      <h1>Hello World</h1>
      <p>Welcome to Vitest with React Testing Library</p>
    </div>
  )
}

describe('Browser Hello World', () => {
  it('should render Hello World component', () => {
    render(<HelloWorld />)
    expect(screen.getByRole('heading', { level: 1, name: 'Hello World' })).toBeDefined()
  })

  it('should render the welcome message', () => {
    render(<HelloWorld />)
    expect(screen.getByText('Welcome to Vitest with React Testing Library')).toBeDefined()
  })

  it('should pass a simple assertion', () => {
    expect(1 + 1).toBe(2)
  })
})

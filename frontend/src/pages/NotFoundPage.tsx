import { Link } from 'react-router-dom'
import { Button } from '@/components/atoms/Button'
import { EmptyState } from '@/components/molecules/Feedback'

export function NotFoundPage() {
  return <div className="not-found"><EmptyState title="Page not found" description="This workspace route does not exist or may have moved." action={<Link to="/connections"><Button variant="primary">Back to connections</Button></Link>} /></div>
}

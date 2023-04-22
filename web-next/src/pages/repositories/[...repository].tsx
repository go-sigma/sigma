import { useRouter } from 'next/router';


export default function Home() {
  const router = useRouter();
  const { repository } = router.query;

  let getRepository = (repository: string | string[] | undefined) => {
    if (typeof repository === 'string') {
      return repository
    } else if (Array.isArray(repository)) {
      return repository.join('/')
    } else {
      return ""
    }
  }

  return <>
    get repository {getRepository(repository)}
  </>
}

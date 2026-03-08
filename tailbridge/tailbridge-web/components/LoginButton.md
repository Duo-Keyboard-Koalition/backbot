// components/LoginButton.tsx

import { signIn, signOut, useSession } from 'next-auth/react';

const LoginButton: React.FC = () => {
  const { data: session } = useSession();

  return (
    <div>
      {!session ? (
        <button onClick={() => signIn('discord')}>Sign in with Discord</button>
      ) : (
        <>
          <span>Welcome, {session.user?.name}!</span>
          <button onClick={() => signOut()}>Sign out</button>
        </>
      )}
    </div>
  );
};

export default LoginButton;

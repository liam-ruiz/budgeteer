import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { map, take } from 'rxjs';
import { AuthService } from '../services/auth';

export const authGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);

  return auth.isLoggedIn$.pipe(
    take(1), // ensures the observable completes after the first check
    map((isLoggedIn) => {
      return isLoggedIn ? true : router.createUrlTree(['/login']);
    })
  );
};
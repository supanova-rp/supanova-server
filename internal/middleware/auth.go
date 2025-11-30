package middleware

import (
	"context"

	"github.com/labstack/echo/v4"
)

type ContextKey string

const UserIDContextKey ContextKey = "userID"

// const nonAdminPaths = [
//   `${routePath}/course`,
//   `${routePath}/assigned-course-titles`,
//   `${routePath}/get-progress`,
//   `${routePath}/update-progress`,
//   `${routePath}/set-intro-completed`,
//   `${routePath}/set-course-completed`,
//   `${routePath}/video-url`,
//   `${routePath}/materials`,
//   `${routePath}/get-quiz-state`,
//   `${routePath}/set-quiz-state`,
//   `${routePath}/increment-attempts`,
// ];

// const verifyUser = async (req, res, next) => {
//   const token = req.body.access_token;

//   if (!token) {
//     return res.status(401).json({ error: 'Unauthorized' });
//   }

//   try {
//     const user = await admin.auth().verifyIdToken(token);

//     if (user) {
//       // If it's a non admin request (e.g. to get courses), then allow request to go through
//       if (nonAdminPaths.includes(req.path)) {
//         req.userId = user.uid; // Set the userId on req for usage in queries
//         req.isAdmin = !!user.admin;

//         return next();
//       }

//       // Otherwise only allow request to go through if user is admin
//       if (user.admin) {
//         req.userId = user.uid; // Set the userId on req for usage in queries
//         req.isAdmin = true;

//         return next();
//       }
//     }

//     return res.status(401).json({ error: 'Unauthorized' });
//   } catch (error) {
//     console.log(error);

//     return res.status(500).json({ error: 'Internal server error' });
//   }
// };

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: implement auth middleware using firebase
		return next(c)
	}
}

func TestAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Request().Header.Get("X-Test-User-ID")

		ctx := context.WithValue(c.Request().Context(), UserIDContextKey, userID)
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}

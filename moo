SELECT
				c.time, u.user, x, y, z, m.material, c.amount, c.action, c.rolled_back
		   	  FROM co_container c
			  LEFT JOIN (co_user u, co_material_map m) on (c.type = m.rowid and c.user = u.rowid)
			  WHERE c.wid = 1
			    AND c.x >= 1 AND c.x <= 4
				AND c.y >= 2 AND c.y <= 5
				AND c.z >= 3 AND c.z <= 6
			  ORDER BY c.time

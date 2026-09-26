package database

import (
 "context"
 "database/sql"
 "math"
 "time"

 "github.com/txyyddss/Remna-User-Panel/internal/activity"
)

// raffleTrafficCoverageTx rejects a seat set whose possible cumulative
// negative traffic grants could make any eligible term nonpositive.
func raffleTrafficCoverageTx(ctx context.Context,tx *sql.Tx,userID string,draw activity.LuckyDraw,seats int,now time.Time) error {
 if seats<=0 { return nil }
 maxDeductionGiB:=int64(0)
 for _,prize:=range draw.Prizes {
  if prize.Reward.Kind==activity.RewardTrafficGrant && prize.Reward.Range!=nil && prize.Reward.Range.Min<0 {
   maxDeductionGiB=max(maxDeductionGiB,-prize.Reward.Range.Min)
  }
 }
 if maxDeductionGiB==0 { return nil }
 _,minimumTraffic,err:=activeRewardPurchase(ctx,tx,userID,now)
 if err!=nil { return err }
 for _,prize:=range draw.Prizes {
  switch prize.Reward.Kind {
  case activity.RewardEntitlementGrant:
   minimumTraffic=min(minimumTraffic,prize.Reward.TrafficLimitBytes)
	case activity.RewardCoreComboSwitch:
		var targetTraffic int64
		if err=tx.QueryRowContext(ctx,`SELECT traffic_limit_bytes FROM combos WHERE id=?`,prize.Reward.ComboID).Scan(&targetTraffic);err!=nil {
			if err==sql.ErrNoRows { return ErrConflict }
			return err
   }
   minimumTraffic=min(minimumTraffic,targetTraffic)
  }
 }
 if maxDeductionGiB>math.MaxInt64/(1<<30) { return activity.ErrInvalidInput }
 perSeat:=maxDeductionGiB*(1<<30)
 if int64(seats)>math.MaxInt64/perSeat || minimumTraffic<=int64(seats)*perSeat { return ErrConflict }
 return nil
}

func validateRaffleTrafficForEntriesTx(ctx context.Context,tx *sql.Tx,draw activity.LuckyDraw,now time.Time) error {
 rows,err:=tx.QueryContext(ctx,`SELECT user_id,COUNT(*) FROM activity_raffle_tickets
  WHERE draw_id=? AND status='active' GROUP BY user_id`,draw.ID)
 if err!=nil { return err }
 type count struct { userID string; seats int }
 counts:=make([]count,0)
 for rows.Next() {
  var item count
  if err=rows.Scan(&item.userID,&item.seats);err!=nil { _=rows.Close();return err }
  counts=append(counts,item)
 }
 if err=rows.Err();err!=nil { _=rows.Close();return err }
 if err=rows.Close();err!=nil { return err }
 for _,item:=range counts {
  if err=raffleTrafficCoverageTx(ctx,tx,item.userID,draw,item.seats,now);err!=nil { return err }
 }
 return nil
}

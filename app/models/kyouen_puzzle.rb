class KyouenPuzzle
  attr_accessor :id, :stage_no, :size, :stage, :creator

  def initialize(entity)
    @id = entity.key.path[0].id
    @stage_no = entity.properties['stageNo'].integer_value
    @size = entity.properties['size'].integer_value
    @stage = entity.properties['stage'].string_value
    @creator = entity.properties['creator'].string_value
  end

  class << self
    def recent(limit = 5)
      client = Datastore::Client.new
      result = client.query(
        'SELECT * FROM KyouenPuzzle ORDER BY stageNo LIMIT @1',
        [Datastore::Parameter.new(limit)]
      )
      return result.map{|r|
        puzzle = KyouenPuzzle.new(r.entity)
      }
    end
  end
end

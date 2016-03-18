class KyouensController < ApplicationController
  before_action :set_kyouen, only: [:show, :update, :destroy]

  # GET /kyouens
  def index
    @kyouens = Kyouen.all

    render json: @kyouens
  end

  # GET /kyouens/1
  def show
    render json: @kyouen
  end

  # POST /kyouens
  def create
    @kyouen = Kyouen.new(kyouen_params)

    if @kyouen.save
      render json: @kyouen, status: :created, location: @kyouen
    else
      render json: @kyouen.errors, status: :unprocessable_entity
    end
  end

  # PATCH/PUT /kyouens/1
  def update
    if @kyouen.update(kyouen_params)
      render json: @kyouen
    else
      render json: @kyouen.errors, status: :unprocessable_entity
    end
  end

  private
    # Use callbacks to share common setup or constraints between actions.
    def set_kyouen
      @kyouen = Kyouen.find(params[:id])
    end

    # Only allow a trusted parameter "white list" through.
    def kyouen_params
      params.require(:kyouen).permit(:size, :stage, :creator)
    end
end
